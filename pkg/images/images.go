package images

import (
	"errors"
	"os"
	"strings"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

var repository string = os.Getenv("REPOSITORY")

func rename(name string) (string, string, string) {
	var registry, img, tag, newName string
	list := strings.Split(name, "/")
	if len(list) == 2 {
		img = list[1]
	} else {
		img = list[0]
	}
	if strings.Contains(img, ":") {
		list := strings.Split(img, ":")
		registry = list[0]
		tag = list[1]
	} else {
		registry = img
	}
	registry = repository + "/" + registry
	if len(tag) != 0 {
		newName = registry + ":" + tag
	} else {
		newName = registry
	}
	return registry, tag, newName
}

func imagePresent(registry, tag string, opt remote.Option) bool {
	rep, _ := name.NewRepository(registry)
	list, _ := remote.List(rep, opt)
	for _, t := range list {
		if t == tag {
			return true
		}
	}
	return false
}

func fetchCredentials() (authn.Authenticator, error) {
	username := os.Getenv("USERNAME")
	password := os.Getenv("PASSWORD")
	if len(username) == 0 || len(password) == 0 {
		return nil, errors.New("Failed to fetch credentials")
	}
	auth := authn.AuthConfig{
		Username: username,
		Password: password,
	}
	authenticator := authn.FromConfig(auth)
	return authenticator, nil
}

// Process public image to retag and push to private registry
func Process(imgName string) (string, error) {
	ref, err := name.ParseReference(imgName)
	if err != nil {
		return "", err
	}
	authenticator, err := fetchCredentials()
	if err != nil {
		return "", err
	}
	opt := remote.WithAuth(authenticator)
	img, err := remote.Image(ref)
	if err != nil {
		return "", err
	}
	registry, tag, newName := rename(imgName)
	newRef, err := name.ParseReference(newName)
	if err != nil {
		return "", err
	}
	if !imagePresent(registry, tag, opt) {
		if err := remote.Write(newRef, img, opt); err != nil {
			return "", err
		}
	}
	return newName, nil
}
