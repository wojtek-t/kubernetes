package main

import (
	"os"
)

func main() {
	// The problem is that the client-go code generator sees ParentReference and thinks it's a top-level resource
	// because we put `+genclient` on it, or because it thinks it needs an applyconfiguration.
	// Oh wait, ApplyConfigurations are generated for all structs used in an API object.
	// `ParentReferenceApplyConfiguration` is not defined. Why?
	// Because ApplyConfigurations generation failed maybe?
	// Oh, `ParentReference` was added manually but maybe without +k8s:deepcopy-gen tag? No, it doesn't need it.

	// Wait! I removed `+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object` from ParentReference earlier,
	// BUT `ParentReference` was still generated as a top-level resource in listers!
	// Why? Because I put it *before* MultiPodGroup which had `+genclient` maybe?
	// Ah, I put `+genclient` on `MultiPodGroup`? No, wait.
}
