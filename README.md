# `Image Clone Controller`

The `Image Clone Controller` watches for Deployments/DaemonSets and caches the images by re-uploading to our own registry repository and reconfiguring the applications to use these copies.

It leans heavily on the lower level Kubernetes [`controller-runtime`](https://github.com/kubernetes-sigs/controller-runtime) package which is a set of go libraries for building
Controllers leveraged by [Kubebuilder](https://book.kubebuilder.io/), 
[Operator SDK](https://github.com/operator-framework/operator-sdk)
and [`remote`](https://github.com/google/go-containerregistry/tree/main/pkg/v1/remote) package which implements a client for accessing a registry,
per the [OCI distribution spec](https://github.com/opencontainers/distribution-spec/blob/master/spec.md).

### Goal
Goal here is to be safe against the risk of public container images disappearing from the registry while
we use them, breaking our deployments.

## Working
` For demo purposes we used newly created "backupregistry" Docker Repositry `
* Watch the Kubernetes Deployment and DaemonSet objects
* Check if any of them provision pods with images that are not from the our registry 
* If yes, copy the image over to the corresponding repository and tag in our registry. 
 ( ` Images are pushed if not already present on Registry` )
* Modify the Deployment/DaemonSet to use the image from our registry.
 ( `Deployment/DaemonSets are only updated when healthy` )
* IMPORTANT: The Deployments and DaemonSets in the kube-system namespace
is ignored!
* Additionaly Deployment/DaemonSets from Controller Namespace are also ignored


## Usage
Create a secret to provide credentials of you registry repository and update config/manager/manager.yaml with the Repository name (env variable `REPOSITORY`) and respective secret name `registry-secret` , credentials `registry-username and registry-passowrd`.
Also, if you are changing controller namespace then update `CONTROLLER_NAMESPACE` env in config/manager/manager.yaml

Create the role, role binding, and service account to grant resource permissions to the Image Clone Operator:
```
$ kubectl create -f config/secrets/secret.yaml // create you own secret.yaml
$ kubectl create -f config/rbac/service_account.yaml
$ kubectl create -f config/rbac/role.yaml
$ kubectl create -f config/rbac/role_binding.yaml
$ kubectl create -f config/manager/manager.yaml
```
## IMPORTANT
For `quay.io` based images set "quay.io/<repo name>" as env variable `REPOSITORY`

## Build
In order to build controller:
```
export  BUNDLE_IMG=<img name>
make docker-build docker-push IMG=$BUNDLE_IMG
```
This will build controller with provided images using Makefile.





