/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package multipodgroup

import (
	"context"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/admission"
	"k8s.io/apiserver/pkg/admission/initializer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	schedulingv1alpha2listers "k8s.io/client-go/listers/scheduling/v1alpha2"

	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// PluginName is a string with the name of the plugin
const PluginName = "MultiPodGroupDepth"

// Register registers a plugin
func Register(plugins *admission.Plugins) {
	plugins.Register(PluginName, func(config io.Reader) (admission.Interface, error) {
		return NewPlugin(), nil
	})
}

// Plugin is an implementation of admission.Interface.
// It looks at all PodGroups and MultiPodGroups to ensure that the
// MultiPodGroup hierarchy does not exceed a depth of 4.
type Plugin struct {
	*admission.Handler
	client kubernetes.Interface
	mpgLister schedulingv1alpha2listers.MultiPodGroupLister
}

var _ admission.ValidationInterface = &Plugin{}
var _ initializer.WantsExternalKubeInformerFactory = &Plugin{}
var _ initializer.WantsExternalKubeClientSet = &Plugin{}

// NewPlugin creates a new Plugin instance
func NewPlugin() *Plugin {
	return &Plugin{
		Handler: admission.NewHandler(admission.Create, admission.Update),
	}
}

// SetExternalKubeClientSet implements the WantsExternalKubeClientSet interface.
func (p *Plugin) SetExternalKubeClientSet(client kubernetes.Interface) {
	p.client = client
}

// SetExternalKubeInformerFactory implements the WantsExternalKubeInformerFactory interface.
func (p *Plugin) SetExternalKubeInformerFactory(f informers.SharedInformerFactory) {
	p.mpgLister = f.Scheduling().V1alpha2().MultiPodGroups().Lister()
	p.SetReadyFunc(f.Scheduling().V1alpha2().MultiPodGroups().Informer().HasSynced)
}

// Validate initialization.
func (p *Plugin) ValidateInitialization() error {
	if p.client == nil {
		return fmt.Errorf("missing client")
	}
	if p.mpgLister == nil {
		return fmt.Errorf("missing mpgLister")
	}
	return nil
}

// Validate makes sure that the object doesn't violate the depth limits.
func (p *Plugin) Validate(ctx context.Context, a admission.Attributes, o admission.ObjectInterfaces) error {
	if a.GetResource().GroupResource() != scheduling.Resource("podgroups") &&
		a.GetResource().GroupResource() != scheduling.Resource("multipodgroups") {
		return nil
	}

	obj := a.GetObject()
	if obj == nil {
		return nil
	}

	namespace := a.GetNamespace()
	var parentRef *scheduling.ParentReference

	switch o := obj.(type) {
	case *scheduling.PodGroup:
		parentRef = o.Spec.ParentRef
	case *scheduling.MultiPodGroup:
		parentRef = o.Spec.ParentRef
	default:
		return nil
	}

	if parentRef == nil {
		return nil
	}

	// Calculate the depth going upwards
	// We count the current object as 1.
	// We allow a maximum depth of 4 (i.e. Root MPG -> MPG -> MPG -> PG/MPG).
	// Let's traverse the parents.
	depth := 1
	currentParentName := parentRef.Name

	for {
		if currentParentName == "" {
			break
		}

		if currentParentName == a.GetName() {
			return admission.NewForbidden(a, fmt.Errorf("circular reference detected for MultiPodGroup %s", a.GetName()))
		}

		mpg, err := p.mpgLister.MultiPodGroups(namespace).Get(currentParentName)
		if err != nil {
			if errors.IsNotFound(err) {
				// Parent doesn't exist yet, we can't fully validate depth but we shouldn't fail
				// admission because objects can be created out of order.
				break
			}
			return admission.NewForbidden(a, fmt.Errorf("error getting parent MultiPodGroup %s: %v", currentParentName, err))
		}

		depth++
		if depth > 4 {
			return admission.NewForbidden(a, fmt.Errorf("MultiPodGroup tree depth exceeds the maximum allowed layers of 4"))
		}

		if mpg.Spec.ParentRef != nil {
			currentParentName = mpg.Spec.ParentRef.Name
		} else {
			break
		}
	}

	return nil
}
