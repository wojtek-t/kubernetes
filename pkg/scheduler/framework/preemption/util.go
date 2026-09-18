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

package preemption

import (
	v1 "k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1"
	schedulingv1alpha3 "k8s.io/api/scheduling/v1alpha3"
	schedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	corev1helpers "k8s.io/component-helpers/scheduling/corev1"
	fwk "k8s.io/kube-scheduler/framework"
	"k8s.io/kube-scheduler/util"
)

// PodTerminatingByPreemption returns true if the pod is in the termination state caused by scheduler preemption.
func PodTerminatingByPreemption(p *v1.Pod) bool {
	if p.DeletionTimestamp == nil {
		return false
	}

	for _, condition := range p.Status.Conditions {
		if condition.Type == v1.DisruptionTarget {
			return condition.Status == v1.ConditionTrue && condition.Reason == v1.PodReasonPreemptionByScheduler
		}
	}
	return false
}

// MoreImportantVictim decides which of two preemption units is considered more critical.
// It applies the following rules in order:
//
//  1. Priority:
//     Higher priority units are more important.
//
//  2. Workload Type:
//     CompositePodGroups are considered more important than PodGroups, and
//     PodGroups are more important than individual Pods to preserve group integrity.
//
//  3. Runtime / Start Time (for individual Pods):
//     For two individual Pods, the one that started earlier (longer runtime)
//     is more important. This honors "first-come, first-served".
//
//  4. Group Size (for PodGroups and CompositePodGroups):
//     If both units are of the same type, the one with more members (larger size) is considered
//     more important. This avoids the high cost of rescheduling massive jobs.
//
//  5. Start Time (Tie-breaker for group types):
//     If group types and sizes are equal, the group that started earlier (has the oldest pod)
//     is more important.
func MoreImportantVictim(vi1, vi2 Victim) bool {
	if vi1.Priority() != vi2.Priority() {
		return vi1.Priority() > vi2.Priority()
	}

	rank1 := victimRank(vi1)
	rank2 := victimRank(vi2)

	if rank1 != rank2 {
		return rank1 > rank2
	}

	if len(vi1.Pods()) != len(vi2.Pods()) {
		return len(vi1.Pods()) > len(vi2.Pods())
	}

	t1, t2 := vi1.EarliestStartTime(), vi2.EarliestStartTime()
	if t1 != nil && t2 != nil && !t1.Equal(t2) {
		return t1.Before(t2)
	}

	return podIdentityLess(vi1, vi2)
}

// podIdentityLess provides a deterministic, stable ordering of victims based on
// the representative pod's UID. The first pod in a victim's list is stable once
// the victim is constructed, so a direct comparison is sufficient.
func podIdentityLess(vi1, vi2 Victim) bool {
	return vi1.Pods()[0].GetPod().UID < vi2.Pods()[0].GetPod().UID
}

func victimRank(vi Victim) int {
	switch vi.Type() {
	case fwk.CompositePodGroupKeyType:
		return 3
	case fwk.PodGroupKeyType:
		return 2
	default:
		return 1
	}
}

// compareVictimStructuralImportance compares two victims on structural importance:
// Priority, Workload Type Rank (CompositePodGroup > PodGroup > Pod), and Group Size.
// Returns > 0 if vi1 is more important, < 0 if vi2 is more important, and 0 if equal.
func compareVictimStructuralImportance(vi1, vi2 Victim) int {
	if vi1.Priority() != vi2.Priority() {
		if vi1.Priority() > vi2.Priority() {
			return 1
		}
		return -1
	}

	rank1 := victimRank(vi1)
	rank2 := victimRank(vi2)
	if rank1 != rank2 {
		if rank1 > rank2 {
			return 1
		}
		return -1
	}

	if len(vi1.Pods()) != len(vi2.Pods()) {
		if len(vi1.Pods()) > len(vi2.Pods()) {
			return 1
		}
		return -1
	}

	return 0
}

// CompareVictimSets compares the preemption cost of two sets of victims (each sorted
// descending by MoreImportantVictim).
// Returns < 0 if set1 is less costly (better to preempt) than set2,
// > 0 if set1 is more costly (worse to preempt) than set2,
// and 0 if both sets have equal preemption cost.
func CompareVictimSets(set1, set2 []*DomainVictim, numPDBViolations1, numPDBViolations2 int) int {
	if len(set1) == 0 && len(set2) == 0 {
		return 0
	}
	if len(set1) == 0 {
		return -1
	}
	if len(set2) == 0 {
		return 1
	}

	if numPDBViolations1 != numPDBViolations2 {
		if numPDBViolations1 < numPDBViolations2 {
			return -1
		}
		return 1
	}

	minLen := min(len(set1), len(set2))
	for i := 0; i < minLen; i++ {
		if cmp := compareVictimStructuralImportance(set1[i], set2[i]); cmp != 0 {
			return cmp
		}
	}

	if len(set1) != len(set2) {
		if len(set1) < len(set2) {
			return -1
		}
		return 1
	}

	var totalPods1, totalPods2 int
	for _, v := range set1 {
		totalPods1 += len(v.Pods())
	}
	for _, v := range set2 {
		totalPods2 += len(v.Pods())
	}
	if totalPods1 != totalPods2 {
		if totalPods1 < totalPods2 {
			return -1
		}
		return 1
	}

	for i := 0; i < minLen; i++ {
		t1, t2 := set1[i].EarliestStartTime(), set2[i].EarliestStartTime()
		if t1 != nil && t2 != nil && !t1.Equal(t2) {
			if t1.After(t2.Time) {
				// set1[i] started more recently (shorter runtime), so it is less costly to preempt.
				return -1
			}
			return 1
		}
	}

	return 0
}

// TraverseHierarchyUp traverses the hierarchy of PodGroups/CompositePodGroups upward from startKey.
// At each node (either a PodGroup or a CompositePodGroup), visitFn is called.
// If visitFn returns (stop = true), the traversal stops immediately.
// If the next parent is missing, or cannot be listed, the traversal also stops.
// This method assumes that both GenericWorkload as well as CompositePodGroup feature gates are enabled.
// TODO: log/return an error if there is a gap in the hierarchy.
func TraverseHierarchyUp(
	namespace string,
	startKey fwk.EntityKey,
	pgLister fwk.PodGroupLister,
	cpgLister fwk.CompositePodGroupLister,
	visitFn func(key fwk.EntityKey, pg *schedulingv1beta1.PodGroup, cpg *schedulingv1alpha3.CompositePodGroup) (stop bool),
) {
	if pgLister == nil || cpgLister == nil {
		return
	}
	currentKey := startKey
	visited := sets.New[fwk.EntityKey]()
	for range schedulingv1beta1.WorkloadMaxTreeDepth {
		if visited.Has(currentKey) {
			break
		}
		visited.Insert(currentKey)

		switch currentKey.Type {
		case fwk.PodGroupKeyType:
			pg, err := pgLister.Get(namespace, currentKey.Name)
			if err != nil || pg == nil {
				return
			}
			if visitFn(currentKey, pg, nil) {
				return
			}
			if pg.Spec.ParentCompositePodGroupName == nil {
				return
			}
			currentKey = fwk.CompositePodGroupKey(namespace, *pg.Spec.ParentCompositePodGroupName)

		case fwk.CompositePodGroupKeyType:
			cpg, err := cpgLister.Get(namespace, currentKey.Name)
			if err != nil || cpg == nil {
				return
			}
			if visitFn(currentKey, nil, cpg) {
				return
			}
			if cpg.Spec.ParentCompositePodGroupName == nil {
				return
			}
			currentKey = fwk.CompositePodGroupKey(namespace, *cpg.Spec.ParentCompositePodGroupName)
		}
	}
}

// GetPodPriority returns the effective preemption priority of a pod. If the pod belongs to
// a pod group or composite pod group hierarchy, it returns the priority of the root group of the hierarchy.
// If podGroupLister is nil or the pod does not belong to a pod group, it returns the pod's own priority.
// If compositePodGroupLister is nil or a parent composite pod group in the hierarchy is not found,
// it falls back to the priority of the last successfully resolved group (or the pod's own priority).
// TODO: log/return an error if there is a gap in the hierarchy.
func GetPodPriority(p *v1.Pod, podGroupLister fwk.PodGroupLister, compositePodGroupLister fwk.CompositePodGroupLister) int32 {
	if p.Spec.SchedulingGroup == nil || podGroupLister == nil {
		return corev1helpers.PodPriority(p)
	}
	if compositePodGroupLister == nil {
		pg, err := podGroupLister.Get(p.Namespace, *p.Spec.SchedulingGroup.PodGroupName)
		if err != nil || pg == nil {
			return corev1helpers.PodPriority(p)
		}
		return util.PodGroupPriority(pg)
	}

	startKey := fwk.PodGroupKey(p.Namespace, *p.Spec.SchedulingGroup.PodGroupName)
	var lastPG *schedulingv1beta1.PodGroup
	var lastCPG *schedulingv1alpha3.CompositePodGroup

	TraverseHierarchyUp(p.Namespace, startKey, podGroupLister, compositePodGroupLister, func(key fwk.EntityKey, pg *schedulingv1beta1.PodGroup, cpg *schedulingv1alpha3.CompositePodGroup) bool {
		if pg != nil {
			lastPG = pg
			lastCPG = nil
		} else if cpg != nil {
			lastCPG = cpg
			lastPG = nil
		}
		return false
	})

	if lastCPG != nil {
		return util.CompositePodGroupPriority(lastCPG)
	}
	if lastPG != nil {
		return util.PodGroupPriority(lastPG)
	}
	return corev1helpers.PodPriority(p)
}

// FilterVictimsWithPDBViolation groups "victims" into "violatingVictims"
// and "nonViolatingVictims" based on whether their PDBs will be violated if they are
// preempted.
func FilterVictimsWithPDBViolation[T Victim](victims []T, pdbs []*policy.PodDisruptionBudget) (violatingVictims []ViolatingVictim[T], nonViolatingVictims []T) {
	pdbsAllowed := make([]int32, len(pdbs))
	podIsViolating := func(pod *v1.Pod) bool {
		for i, pdb := range pdbs {
			if pdb.Namespace != pod.Namespace {
				continue
			}
			selector, err := metav1.LabelSelectorAsSelector(pdb.Spec.Selector)
			if err != nil {
				// This object has an invalid selector, it does not match the pod
				continue
			}
			if !selector.Matches(labels.Set(pod.Labels)) {
				continue
			}

			// Existing in DisruptedPods means it has been processed in API server,
			// we don't treat it as a violating case.
			if _, exist := pdb.Status.DisruptedPods[pod.Name]; exist {
				continue
			}
			// Only decrement the matched pdb when it's not in its <DisruptedPods>;
			// otherwise we may over-decrement the budget number.
			pdbsAllowed[i]--
			// We have found a matching PDB.
			if pdbsAllowed[i] < 0 {
				return true
			}
		}

		return false
	}

	for i, pdb := range pdbs {
		pdbsAllowed[i] = pdb.Status.DisruptionsAllowed
	}

	// A Victim is atomic: if any single pod inside it would violate a PDB on
	// eviction, the entire Victim is classified as violating. This is required
	// for PodGroup victims, where partial preemption is not allowed — we cannot
	// "reprieve" only the pods that violate a PDB and keep the rest evicted.
	for _, victim := range victims {
		violatingPods := 0

		for _, pi := range victim.Pods() {
			if podIsViolating(pi.GetPod()) {
				violatingPods++
			}
		}
		if violatingPods > 0 {
			violatingVictims = append(violatingVictims, ViolatingVictim[T]{
				Victim:       victim,
				ViolateCount: violatingPods,
			})
		} else {
			nonViolatingVictims = append(nonViolatingVictims, victim)
		}
	}

	return violatingVictims, nonViolatingVictims
}
