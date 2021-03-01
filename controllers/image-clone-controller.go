/*
Copyright 2021.

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
	"os"
	"strings"

	"github.com/go-logr/logr"
	"github.com/imharshita/image-controller/pkg/images"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// KubeNs Namespace to exclude in Reconiler
	KubeNs       = "kube-system"
	ControllerNs = "system"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// DaemonSetReconciler reconciles a DaemonSet object
type DaemonSetReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func isImagePresent(image string) bool {
	Registry := os.Getenv("REPOSITORY")
	if len(Registry) == 0 {
		return false
	} else if !strings.HasPrefix(image, Registry){
		return true
	}
	return false
}

func isDaemonSetReady(daemonsets *appsv1.DaemonSet) bool {
	status := daemonsets.Status
	desired := status.DesiredNumberScheduled
	ready := status.NumberReady
	if desired == ready && desired > 0 {
		return true
	}
	return false
}

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("execution")
)

// Reconcile recociles DaemonSet to uodate the image
func (r *DaemonSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqNamespace := req.NamespacedName.Namespace
	if reqNamespace != KubeNs && reqNamespace != ControllerNs {
		r.Log.WithValues("daemonset", req.NamespacedName)
		daemonsets := &appsv1.DaemonSet{}
		err := r.Get(context.TODO(), req.NamespacedName, daemonsets)
		if err != nil {
			return reconcile.Result{}, err
		}
		if isDaemonSetReady(daemonsets) {
			containers := daemonsets.Spec.Template.Spec.Containers
			for i, c := range containers {
				if isImagePresent(c.Image) {
					var msg string
					msg = fmt.Sprintf("Retagging image %s of daemonset: %s", c.Image, daemonsets.Name)
					setupLog.Info(msg)
					img, err := images.Process(c.Image)
					if err != nil {
						msg = fmt.Sprintf("Failed to process image: %s", img)
						setupLog.Error(err, msg)
						return reconcile.Result{}, err
					}
					// update image
					msg = fmt.Sprintf("Updating image %s of daemonset: %s", c.Image, daemonsets.Name)
					setupLog.Info(msg)
					daemonsets.Spec.Template.Spec.Containers[i].Image = img
					err = r.Update(context.TODO(), daemonsets)
					if err != nil {
						return reconcile.Result{}, err
					}
					msg = fmt.Sprintf("Updated image: %s -> %s", c.Image, img)
					setupLog.Info(msg)
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

func isDeploymentReady(deployments *appsv1.Deployment) bool {
	status := deployments.Status
	desired := status.Replicas
	ready := status.ReadyReplicas
	if desired == ready && desired > 0 {
		return true
	}
	return false
}

// Reconcile reconciles Deployment to update the image
func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqNamespace := req.NamespacedName.Namespace
	if reqNamespace != KubeNs && reqNamespace != ControllerNs {
		r.Log.WithValues("deployment", req.NamespacedName)
		deployments := &appsv1.Deployment{}
		err := r.Get(context.TODO(), req.NamespacedName, deployments)
		if err != nil {
			return reconcile.Result{}, err
		}
		if isDeploymentReady(deployments) {
			containers := deployments.Spec.Template.Spec.Containers
			for i, c := range containers {
				if isImagePresent(c.Image) {
					var msg string
					msg = fmt.Sprintf("Retagging image %s of deployment: %s", c.Image, deployments.Name)
					setupLog.Info(msg)
					img, err := images.Process(c.Image)
					if err != nil {
						return reconcile.Result{}, err
					}
					// Update image
					msg = fmt.Sprintf("Updating image %s of deployment: %s", c.Image, deployments.Name)
					setupLog.Info(msg)
					deployments.Spec.Template.Spec.Containers[i].Image = img
					err = r.Update(context.TODO(), deployments)
					if err != nil {
						return reconcile.Result{}, err
					}
					msg = fmt.Sprintf("Updated image: %s -> %s", c.Image, img)
					setupLog.Info(msg)
				}
			}
		}
	}
	return reconcile.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Owns(&corev1.Pod{}).
		Watches(&source.Kind{Type: &appsv1.Deployment{}},
			&handler.EnqueueRequestForObject{}).
		Complete(r)
}

// SetupWithManager sets up the controller with the Manager.
func (r *DaemonSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.DaemonSet{}).
		Owns(&corev1.Pod{}).
		Watches(&source.Kind{Type: &appsv1.DaemonSet{}},
			&handler.EnqueueRequestForObject{}).
		Complete(r)
}
