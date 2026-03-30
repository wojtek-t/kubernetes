package v1beta1

import (
	"k8s.io/apimachinery/pkg/conversion"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

// Convert_scheduling_PodGroupSchedulingPolicy_To_v1beta1_PodGroupSchedulingPolicy drops GangMultiPodGroup since it doesn't exist in v1beta1.
func Convert_scheduling_PodGroupSchedulingPolicy_To_v1beta1_PodGroupSchedulingPolicy(in *scheduling.PodGroupSchedulingPolicy, out *schedulingv1beta1.PodGroupSchedulingPolicy, s conversion.Scope) error {
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
	// Note: GangMultiPodGroup is dropped silently for v1beta1 as it's not supported in that API version.
	return nil
}

func Convert_scheduling_PodGroupTemplate_To_v1beta1_PodGroupTemplate(in *scheduling.PodGroupTemplate, out *schedulingv1beta1.PodGroupTemplate, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupTemplate_To_v1beta1_PodGroupTemplate(in, out, s)
}
