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

package v1alpha2

import (
	"k8s.io/apimachinery/pkg/conversion"
	schedulingv1alpha2 "k8s.io/api/scheduling/v1alpha2"
		"k8s.io/kubernetes/pkg/apis/scheduling"
)

func Convert_v1alpha2_PodGroupSpec_To_scheduling_PodGroupSpec(in *schedulingv1alpha2.PodGroupSpec, out *scheduling.PodGroupSpec, s conversion.Scope) error {
	return autoConvert_v1alpha2_PodGroupSpec_To_scheduling_PodGroupSpec(in, out, s)
}

func Convert_scheduling_PodGroupTemplate_To_v1alpha2_PodGroupTemplate(in *scheduling.PodGroupTemplate, out *schedulingv1alpha2.PodGroupTemplate, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupTemplate_To_v1alpha2_PodGroupTemplate(in, out, s)
}



func Convert_v1alpha2_ParentReference_To_scheduling_ParentReference(in *schedulingv1alpha2.ParentReference, out *scheduling.ParentReference, s conversion.Scope) error {
	out.Name = in.Name
	return nil
}

func Convert_scheduling_ParentReference_To_v1alpha2_ParentReference(in *scheduling.ParentReference, out *schedulingv1alpha2.ParentReference, s conversion.Scope) error {
	out.Name = in.Name
	return nil
}

func Convert_v1alpha2_MultiPodGroupSchedulingPolicy_To_scheduling_MultiPodGroupSchedulingPolicy(in *schedulingv1alpha2.MultiPodGroupSchedulingPolicy, out *scheduling.MultiPodGroupSchedulingPolicy, s conversion.Scope) error {
	return autoConvert_v1alpha2_MultiPodGroupSchedulingPolicy_To_scheduling_MultiPodGroupSchedulingPolicy(in, out, s)
}

func Convert_scheduling_MultiPodGroupSchedulingPolicy_To_v1alpha2_MultiPodGroupSchedulingPolicy(in *scheduling.MultiPodGroupSchedulingPolicy, out *schedulingv1alpha2.MultiPodGroupSchedulingPolicy, s conversion.Scope) error {
	return autoConvert_scheduling_MultiPodGroupSchedulingPolicy_To_v1alpha2_MultiPodGroupSchedulingPolicy(in, out, s)
}
