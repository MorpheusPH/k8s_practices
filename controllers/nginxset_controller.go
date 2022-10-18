/*
Copyright 2022.

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

package controllers

import (
	"context"
	"fmt"

	"github.com/MorpheusPH/nginxcontroller/api/v1beta1"
	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// NginxSetReconciler reconciles a NginxSet object
type NginxSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=api.morpheusph.io,resources=nginxsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.morpheusph.io,resources=nginxsets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.morpheusph.io,resources=nginxsets/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NginxSet object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NginxSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)
	log := r.Log.WithValues("NginxSet", req.NamespacedName)
	nginxset := &v1beta1.NginxSet{}
	if err := r.Get(ctx, req.NamespacedName, nginxset); err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted
			// return and don't requeue
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if !controllerutil.ContainsFinalizer(nginxset, v1beta1.NginxSetFinalizer) {
		patch := client.MergeFrom(nginxset.DeepCopy())
		controllerutil.AddFinalizer(nginxset, v1beta1.NginxSetFinalizer)
		if err := r.Patch(ctx, nginxset, patch); err != nil {
			log.Error(err, "unable to register finalizer")
			return ctrl.Result{}, err
		}
	}

	// examine DeletionTimestamp to determine if object is under deletion
	if !nginxset.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, nginxset)
	}

	if nginxset.Status.ObservedGeneration != nginxset.Generation {
		nginxset.Status.ObservedGeneration = nginxset.Generation
		releaseProgressing(nginxset)
		if updateStatusErr := r.patchStatus(ctx, nginxset); updateStatusErr != nil {
			log.Error(updateStatusErr, "unable to update status after generation update")
			return ctrl.Result{Requeue: true}, updateStatusErr
		}
	}

	ready := findCondition(nginxset, v1beta1.Ready)
	if ready != nil {
		switch ready.Status {
		case metav1.ConditionTrue:
			return ctrl.Result{}, nil
		default:
			if nginxset.Status.Failures >= 5 {
				errorMsg := fmt.Sprintf("%s(%s)", "exceeded the maximum number of release attempts", nginxset.FindCondition(v1beta1.Ready).Message)
				releaseNotReady(nginxset, errorMsg)
				if updateStatusErr := r.patchStatus(ctx, nginxset); updateStatusErr != nil {
					log.Error(updateStatusErr, "unable to update status after reconciliation")
					return ctrl.Result{Requeue: true}, updateStatusErr
				}
				return ctrl.Result{}, nil
			}
		}
	}

	log.Info("reconciling nginxset")

	result, reconcileErr := r.reconcile(ctx, nginxset)
	if reconcileErr != nil {
		releaseNotReady(nginxset, reconcileErr.Error())
		addFailureCounts(nginxset)
	} else {
		setReadyCondition(nginxset, v1beta1.SuccessdedReason)
	}

	if updateStatusErr := r.patchStatus(ctx, nginxset); updateStatusErr != nil {
		log.Error(updateStatusErr, "unable to update status after reconciliation")
		return ctrl.Result{Requeue: true}, updateStatusErr
	}

	return result, reconcileErr
}

func (r *NginxSetReconciler) reconcile(ctx context.Context, nginxset *v1beta1.NginxSet) (ctrl.Result, error) {

	replicas := nginxset.Spec.Replicas

	key := client.ObjectKey{
		Namespace: nginxset.GetNamespace(),
		Name:      nginxset.GetName(),
	}

	current := &appsv1.Deployment{}
	// deployment := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      nginxset.GetName(),
	// 		Namespace: nginxset.GetNamespace(),
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Replicas: replicas,
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: map[string]string{
	// 				"app": "nginx",
	// 			},
	// 		},
	// 		Template: corev1.PodTemplateSpec{
	// 			ObjectMeta: metav1.ObjectMeta{
	// 				Labels: map[string]string{
	// 					"app": "nginx",
	// 				},
	// 			},
	// 			Spec: corev1.PodSpec{
	// 				Containers: []corev1.Container{
	// 					{
	// 						Name:  "web",
	// 						Image: "nginx:1.12",
	// 						Ports: []corev1.ContainerPort{
	// 							{
	// 								Name:          "http",
	// 								Protocol:      corev1.ProtocolTCP,
	// 								ContainerPort: 80,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }
	if err := r.Get(ctx, key, current); err != nil {
		if !apierrors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
		// create deployment object
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      nginxset.GetName(),
				Namespace: nginxset.GetNamespace(),
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: replicas,
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "nginx",
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "nginx",
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "web",
								Image: "nginx:1.12",
								Ports: []corev1.ContainerPort{
									{
										Name:          "http",
										Protocol:      corev1.ProtocolTCP,
										ContainerPort: 80,
									},
								},
							},
						},
					},
				},
			},
		}
		controllerutil.SetOwnerReference(nginxset, deployment, r.Scheme)
		if err := r.Create(ctx, deployment); err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{}, nil
	}

	//patch replicas
	patch := client.MergeFrom(current.DeepCopy())
	current.Spec.Replicas = replicas
	if err := r.Patch(ctx, current, patch); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NginxSetReconciler) reconcileDelete(ctx context.Context, app *v1beta1.NginxSet) (ctrl.Result, error) {
	log := r.Log.WithValues("NginxSet", client.ObjectKeyFromObject(app))
	log.Info("reconcileDelete...")

	// Remove our finalizer from the list
	controllerutil.RemoveFinalizer(app, v1beta1.NginxSetFinalizer)
	if err := r.Update(ctx, app); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NginxSetReconciler) patchStatus(ctx context.Context, obj client.Object) error {
	namespaceName := client.ObjectKeyFromObject(obj)
	current := obj.DeepCopyObject().(client.Object)
	if err := r.Get(ctx, namespaceName, current); err != nil {
		return err
	}

	return r.Status().Patch(ctx, obj, client.MergeFrom(current))
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.NginxSet{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 2,
		}).
		WithEventFilter(predicate.Or(predicate.GenerationChangedPredicate{}, predicate.AnnotationChangedPredicate{})).
		Complete(r)
}
