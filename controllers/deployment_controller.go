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

	"github.com/go-logr/logr"
	"strings"

	"github.com/google/go-containerregistry/pkg/crane"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
	//"github.com/imharshita/image-controller/pkg/images/"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var privateRegistry string = "harshitadocker"

func rename(name string) string {
	image := strings.Split(name, ":")
	img, version := image[0], image[1]
	newName := privateRegistry + "/" + img + ":" + version
	return newName
}

func retag(imgName string) (name.Tag, error) {
	tag, err := name.NewTag(imgName)
	if err != nil {
		return name.Tag{}, err
	}
	return tag, nil
}

func Process(imgName string) (string, error) {
	img, err := crane.Pull(imgName)
	if err != nil {
		return "", err
	}
	newName := rename(imgName)
	tag, err := retag(newName)
	if err != nil {
		return "", err
	}

	_, err = daemon.Write(tag, img)
	if err != nil {
		return "", err
	}
	if err := crane.Push(img, tag.String()); err != nil {
		return "", err
	}
	return newName, nil
}

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// fmt.Println(req.NamespacedName.Namespace)
	reqNamespace := req.NamespacedName.Namespace
	if reqNamespace != "kube-system" {
		_ = r.Log.WithValues("deployment", req.NamespacedName)

		deployments := &appsv1.Deployment{}
		err := r.Get(context.TODO(), req.NamespacedName, deployments)
		if err != nil {
			return reconcile.Result{}, err
		}
		// watch namespace
		//namespaces := deployments.Namespace
		//fmt.Println(namespaces)
		containers := deployments.Spec.Template.Spec.Containers
		for i, c := range containers {
			fmt.Println(c.Image)
			//harshita/nginx:1.14.2
			//string.Prexix()
			//backupregistry/nginx:1.14.2
			if c.Image == "nginx:1.14.2" {
				img, err := Process(c.Image)
				if err != nil {
					return ctrl.Result{}, err
				}
				fmt.Println(img)
				// Update the Deployment
				deployments.Spec.Template.Spec.Containers[i].Image = img
				err = r.Update(context.TODO(), deployments)
				if err != nil {
					return reconcile.Result{}, err
				}
			}

		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
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
