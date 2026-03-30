package main

import (
	"os"
	"strings"
)

func main() {
	path := "plugin/pkg/admission/scheduling/multipodgroup/admission_test.go"
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	content := string(contentBytes)

	// Since we changed the test to use scheduling.MultiPodGroup, we need to populate
	// the store with schedulingv1alpha2.MultiPodGroup because the client-go fake
	// uses external types! Wait, informerFactory.Scheduling().V1alpha2().MultiPodGroups()
	// stores v1alpha2 objects.
	// Oh!
	// In the test, we add `scheduling.MultiPodGroup` into the informer?
	// if mpg, ok := obj.(*scheduling.MultiPodGroup); ok {
	// Let's see what the test has now.
}
