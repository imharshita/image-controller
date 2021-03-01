package images_test

import (
	"testing"

	"github.com/imharshita/image-controller/pkg/images"
)

func TestProcess(t *testing.T) {
	result, _ := images.Process("harshitadocker/nginx:1.14.2")
	if result != "backupregistry/nginx:1.14.2" {
		t.Errorf("Process(\"harshitadocker/nginx:1.14.2\") failed, expected %v, got %v", "backupregistry/nginx:1.14.2", result)
	}
}
