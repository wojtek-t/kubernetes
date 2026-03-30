package v1alpha2

import (
	"k8s.io/apimachinery/pkg/conversion"
	schedulingv1alpha2 "k8s.io/api/scheduling/v1alpha2"
	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/scheduling"
)

func Convert_v1alpha2_PodGroupSpec_To_scheduling_PodGroupSpec(in *schedulingv1alpha2.PodGroupSpec, out *scheduling.PodGroupSpec, s conversion.Scope) error {
	return autoConvert_v1alpha2_PodGroupSpec_To_scheduling_PodGroupSpec(in, out, s)
}

func Convert_scheduling_PodGroupTemplate_To_v1alpha2_PodGroupTemplate(in *scheduling.PodGroupTemplate, out *schedulingv1alpha2.PodGroupTemplate, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupTemplate_To_v1alpha2_PodGroupTemplate(in, out, s)
}

func Convert_v1alpha2_TypedLocalObjectReference_To_core_TypedLocalObjectReference(in *schedulingv1alpha2.TypedLocalObjectReference, out *core.TypedLocalObjectReference, s conversion.Scope) error {
	out.APIGroup = (*string)(&in.APIGroup)
	out.Kind = in.Kind
	out.Name = in.Name
	return nil
}

func Convert_core_TypedLocalObjectReference_To_v1alpha2_TypedLocalObjectReference(in *core.TypedLocalObjectReference, out *schedulingv1alpha2.TypedLocalObjectReference, s conversion.Scope) error {
	if in.APIGroup != nil {
		out.APIGroup = *in.APIGroup
	}
	out.Kind = in.Kind
	out.Name = in.Name
	return nil
}

// Convert_v1alpha2_PodGroupSchedulingPolicy_To_scheduling_PodGroupSchedulingPolicy handles the new field GangMultiPodGroup.
func Convert_v1alpha2_PodGroupSchedulingPolicy_To_scheduling_PodGroupSchedulingPolicy(in *schedulingv1alpha2.PodGroupSchedulingPolicy, out *scheduling.PodGroupSchedulingPolicy, s conversion.Scope) error {
	return autoConvert_v1alpha2_PodGroupSchedulingPolicy_To_scheduling_PodGroupSchedulingPolicy(in, out, s)
}

// Convert_scheduling_PodGroupSchedulingPolicy_To_v1alpha2_PodGroupSchedulingPolicy handles the new field GangMultiPodGroup.
func Convert_scheduling_PodGroupSchedulingPolicy_To_v1alpha2_PodGroupSchedulingPolicy(in *scheduling.PodGroupSchedulingPolicy, out *schedulingv1alpha2.PodGroupSchedulingPolicy, s conversion.Scope) error {
	return autoConvert_scheduling_PodGroupSchedulingPolicy_To_v1alpha2_PodGroupSchedulingPolicy(in, out, s)
}
