# `Image Clone Controller`

The `Image Clone Controller` watches for Deployments/DaemonSets and reconciles if images used are not from
registry "backupregistry"

It leans heavily on the lower level Kubernetes [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime) package which is a set of go libraries for building
Controllers. It is leveraged by [Kubebuilder](https://book.kubebuilder.io/) and
[Operator SDK](https://github.com/operator-framework/operator-sdk).
and [`remote`](https://github.com/google/go-containerregistry/tree/main/pkg/v1/remote) package implements a client for accessing a registry,
per the [OCI distribution spec](https://github.com/opencontainers/distribution-spec/blob/master/spec.md).

Controller is written in Go

* Watch the Kubernetes Deployment and DaemonSet objects
* Check if any of them provision pods with images that are not from the backup
registry
* If yes, copy the image over to a corresponding repository and tag in the backup
registry
* Modify the Deployment/DaemonSet to use the image from the backup registry
* IMPORTANT: The Deployments and DaemonSets in the kube-system namespace
is ignored!

## Working 
Controller watches for Deployments/DaemonSets assuming if images are not from "backupregistry" than they are public. Its important in order to access the images
Then images are retagged to be pushed to "backupregistry" and ultimately are pushed if already not present there

## Usage
Create the role, role binding, and service account to grant resource permissions to the Operator, and Image Clone Operator:
```
$ kubectl create -f conifg/secrets/secret.yaml
$ kubectl create -f conifg/rbac/service_account.yaml
$ kubectl create -f conifg/rbac/role.yaml
$ kubectl create -f conifg/rbac/role_binding.yaml
$ kubectl create -f config/controllers/image-clone-controller.yaml
```
## IMPORTANT
In order to use controller a secret file named "registry-secret" must be present 

## Build
In order to build controller:

```
export  BUNDLE_IMG=<img name>
make docker-build docker-push IMG=$BUNDLE_IMG
```
### Goal
Goal here is to be safe against the risk of public container images disappearing from the registry while
we use them, breaking our deployments.




