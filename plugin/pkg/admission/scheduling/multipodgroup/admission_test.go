package multipodgroup

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/admission"
	clientgofake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/informers"

	"k8s.io/kubernetes/pkg/apis/core"
	"k8s.io/kubernetes/pkg/apis/scheduling"
	schedulingv1alpha2 "k8s.io/api/scheduling/v1alpha2"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		target      runtime.Object
		existing    []runtime.Object
		expectedErr bool
	}{
		{
			name: "no parent, depth 1",
			target: &scheduling.PodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "pg-1", Namespace: "default"},
				Spec: scheduling.PodGroupSpec{},
			},
			existing:    nil,
			expectedErr: false,
		},
		{
			name: "1 parent, depth 2",
			target: &scheduling.PodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "pg-1", Namespace: "default"},
				Spec: scheduling.PodGroupSpec{
					ParentRef: &core.TypedLocalObjectReference{Name: "mpg-1"},
				},
			},
			existing: []runtime.Object{
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-1", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{},
				},
			},
			expectedErr: false,
		},
		{
			name: "3 parents, depth 4",
			target: &scheduling.PodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "pg-1", Namespace: "default"},
				Spec: scheduling.PodGroupSpec{
					ParentRef: &core.TypedLocalObjectReference{Name: "mpg-1"},
				},
			},
			existing: []runtime.Object{
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-1", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-2"},
					},
				},
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-2", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-3"},
					},
				},
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-3", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{},
				},
			},
			expectedErr: false,
		},
		{
			name: "4 parents, depth 5 (should fail)",
			target: &scheduling.PodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "pg-1", Namespace: "default"},
				Spec: scheduling.PodGroupSpec{
					ParentRef: &core.TypedLocalObjectReference{Name: "mpg-1"},
				},
			},
			existing: []runtime.Object{
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-1", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-2"},
					},
				},
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-2", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-3"},
					},
				},
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-3", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-4"},
					},
				},
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-4", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{},
				},
			},
			expectedErr: true,
		},
		{
			name: "circular reference, should detect depth 5 and fail",
			target: &scheduling.MultiPodGroup{
				ObjectMeta: metav1.ObjectMeta{Name: "mpg-1", Namespace: "default"},
				Spec: scheduling.MultiPodGroupSpec{
					ParentRef: &core.TypedLocalObjectReference{Name: "mpg-2"},
				},
			},
			existing: []runtime.Object{
				&schedulingv1alpha2.MultiPodGroup{
					ObjectMeta: metav1.ObjectMeta{Name: "mpg-2", Namespace: "default"},
					Spec: schedulingv1alpha2.MultiPodGroupSpec{
						ParentRef: &schedulingv1alpha2.TypedLocalObjectReference{Name: "mpg-1"},
					},
				},
			},
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := clientgofake.NewSimpleClientset(tt.existing...)
			informerFactory := informers.NewSharedInformerFactory(client, 0)

			for _, obj := range tt.existing {
				if mpg, ok := obj.(*schedulingv1alpha2.MultiPodGroup); ok {
					informerFactory.Scheduling().V1alpha2().MultiPodGroups().Informer().GetStore().Add(mpg)
				}
			}

			plugin := NewPlugin()
			plugin.SetExternalKubeClientSet(client)
			plugin.SetExternalKubeInformerFactory(informerFactory)
			if err := plugin.ValidateInitialization(); err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			gvr := schedulingv1alpha2.SchemeGroupVersion.WithResource("podgroups")
			if _, ok := tt.target.(*scheduling.MultiPodGroup); ok {
				gvr = schedulingv1alpha2.SchemeGroupVersion.WithResource("multipodgroups")
			}

			var targetName string
			if pg, ok := tt.target.(*scheduling.PodGroup); ok {
				targetName = pg.Name
			} else if mpg, ok := tt.target.(*scheduling.MultiPodGroup); ok {
				targetName = mpg.Name
			}

			attrs := admission.NewAttributesRecord(
				tt.target,
				nil,
				schedulingv1alpha2.SchemeGroupVersion.WithKind("PodGroup"),
				"default",
				targetName,
				gvr,
				"",
				admission.Create,
				nil,
				false,
				nil,
			)

			err := plugin.Validate(context.TODO(), attrs, nil)
			if (err != nil) != tt.expectedErr {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}
		})
	}
}
