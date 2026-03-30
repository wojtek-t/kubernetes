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

package v1beta1

import (
	"k8s.io/apimachinery/pkg/conversion"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

func Convert_scheduling_PodGroupSchedulingPolicy_To_v1beta1_PodGroupSchedulingPolicy(in *scheduling.PodGroupSchedulingPolicy, out *schedulingv1beta1.PodGroupSchedulingPolicy, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupSchedulingPolicy_To_v1beta1_PodGroupSchedulingPolicy(in, out, s)
}

func Convert_scheduling_PodGroupTemplate_To_v1beta1_PodGroupTemplate(in *scheduling.PodGroupTemplate, out *schedulingv1beta1.PodGroupTemplate, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupTemplate_To_v1beta1_PodGroupTemplate(in, out, s)
}

func Convert_scheduling_PodGroupSpec_To_v1beta1_PodGroupSpec(in *scheduling.PodGroupSpec, out *schedulingv1beta1.PodGroupSpec, s conversion.Scope) error {
	// Drop ParentRef since it doesn't exist in v1beta1
	return autoConvert_scheduling_PodGroupSpec_To_v1beta1_PodGroupSpec(in, out, s)
}

func Convert_v1beta1_PodGroupSchedulingPolicy_To_scheduling_MultiPodGroupSchedulingPolicy(in *schedulingv1beta1.PodGroupSchedulingPolicy, out *scheduling.MultiPodGroupSchedulingPolicy, s conversion.Scope) error {
	if in.Basic != nil {
		out.Basic = &scheduling.BasicSchedulingPolicy{}
	} else {
		out.Basic = nil
	}
	if in.Gang != nil {
		out.Gang = &scheduling.GangSchedulingPolicy{}
	} else {
		out.Gang = nil
	}
	return nil
}

func Convert_scheduling_MultiPodGroupSchedulingPolicy_To_v1beta1_PodGroupSchedulingPolicy(in *scheduling.MultiPodGroupSchedulingPolicy, out *schedulingv1beta1.PodGroupSchedulingPolicy, s conversion.Scope) error {
	if in.Basic != nil {
		out.Basic = &schedulingv1beta1.BasicSchedulingPolicy{}
	} else {
		out.Basic = nil
	}
	if in.Gang != nil {
		out.Gang = &schedulingv1beta1.GangSchedulingPolicy{}
	} else {
		out.Gang = nil
	}
	return nil
}
