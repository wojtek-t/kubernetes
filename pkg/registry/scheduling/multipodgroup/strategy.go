package multipodgroup

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/apiserver/pkg/registry/generic"
	"k8s.io/apiserver/pkg/storage"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	"sigs.k8s.io/structured-merge-diff/v6/fieldpath"
)

// multiPodGroupStrategy implements behavior for MultiPodGroups
type multiPodGroupStrategy struct {
	runtime.ObjectTyper
	names.NameGenerator
}

// Strategy is the default logic that applies when creating and updating MultiPodGroup
// objects via the REST API.
var Strategy = multiPodGroupStrategy{legacyscheme.Scheme, names.SimpleNameGenerator}

// NamespaceScoped is true for multi pod groups.
func (multiPodGroupStrategy) NamespaceScoped() bool {
	return true
}

// PrepareForCreate clears fields that are not allowed to be set by end users on creation.
func (multiPodGroupStrategy) PrepareForCreate(ctx context.Context, obj runtime.Object) {
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update.
func (multiPodGroupStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
}

// Validate validates a new multi pod group.
func (multiPodGroupStrategy) Validate(ctx context.Context, obj runtime.Object) field.ErrorList {
	mpg := obj.(*scheduling.MultiPodGroup)
	return ValidateMultiPodGroup(mpg)
}

// WarningsOnCreate returns warnings for the creation of the given object.
func (multiPodGroupStrategy) WarningsOnCreate(ctx context.Context, obj runtime.Object) []string {
	return nil
}

// Canonicalize normalizes the object after validation.
func (multiPodGroupStrategy) Canonicalize(obj runtime.Object) {
}

// AllowCreateOnUpdate is false for multi pod groups.
func (multiPodGroupStrategy) AllowCreateOnUpdate() bool {
	return false
}

// ValidateUpdate is the default update validation for an end user.
func (multiPodGroupStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	mpg := obj.(*scheduling.MultiPodGroup)
	oldMpg := old.(*scheduling.MultiPodGroup)
	return ValidateMultiPodGroupUpdate(mpg, oldMpg)
}

// WarningsOnUpdate returns warnings for the given update.
func (multiPodGroupStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

// AllowUnconditionalUpdate is the default update policy for multi pod group objects.
func (multiPodGroupStrategy) AllowUnconditionalUpdate() bool {
	return true
}

// SelectableFields returns a field set that represents the object.
func SelectableFields(mpg *scheduling.MultiPodGroup) fields.Set {
	return generic.ObjectMetaFieldsSet(&mpg.ObjectMeta, true)
}

// GetAttrs returns labels.Set, fields.Set, and error in case the given runtime.Object is not a MultiPodGroup
func GetAttrs(obj runtime.Object) (labels.Set, fields.Set, error) {
	mpg, ok := obj.(*scheduling.MultiPodGroup)
	if !ok {
		return nil, nil, fmt.Errorf("given object is not a MultiPodGroup")
	}
	return labels.Set(mpg.ObjectMeta.Labels), SelectableFields(mpg), nil
}

// MatchMultiPodGroup is the filter used by the generic etcd backend to watch events
// from etcd to clients of the apiserver only interested in specific labels/fields.
func MatchMultiPodGroup(label labels.Selector, field fields.Selector) storage.SelectionPredicate {
	return storage.SelectionPredicate{
		Label:    label,
		Field:    field,
		GetAttrs: GetAttrs,
	}
}

// StatusStrategy implements behavior for MultiPodGroup status updates.
type statusStrategy struct {
	multiPodGroupStrategy
}

// StatusStrategy is the default logic that applies when creating and updating MultiPodGroup
// objects via the REST API.
var StatusStrategy = statusStrategy{Strategy}

// GetResetFields returns the set of fields that get reset by the strategy
// and should not be modified by the user.
func (statusStrategy) GetResetFields() map[fieldpath.APIVersion]*fieldpath.Set {
	return map[fieldpath.APIVersion]*fieldpath.Set{
		"scheduling.k8s.io/v1alpha2": fieldpath.NewSet(
			fieldpath.MakePathOrDie("spec"),
		),
	}
}

// PrepareForUpdate clears fields that are not allowed to be set by end users on update of status
func (statusStrategy) PrepareForUpdate(ctx context.Context, obj, old runtime.Object) {
	newMpg := obj.(*scheduling.MultiPodGroup)
	oldMpg := old.(*scheduling.MultiPodGroup)
	newMpg.Spec = oldMpg.Spec
}

// ValidateUpdate is the default update validation for an end user updating status
func (statusStrategy) ValidateUpdate(ctx context.Context, obj, old runtime.Object) field.ErrorList {
	return field.ErrorList{}
}

// WarningsOnUpdate returns warnings for the given update.
func (statusStrategy) WarningsOnUpdate(ctx context.Context, obj, old runtime.Object) []string {
	return nil
}

func ValidateMultiPodGroup(mpg *scheduling.MultiPodGroup) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}

func ValidateMultiPodGroupUpdate(mpg, oldMpg *scheduling.MultiPodGroup) field.ErrorList {
	allErrs := field.ErrorList{}
	return allErrs
}
