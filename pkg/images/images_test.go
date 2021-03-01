package images_test

import (
	"testing"

	"github.com/imharshita/image-controller/pkg/images"
)

func TestProcess(t *testing.T) {
	result1, err := images.Process("busybox")
	if result1 != "backupregistry/busybox" {
		t.Errorf("Process(\"busybox\") failed, expected %v, got %v  error %v  ", "backupregistry/busybox", result1, err)
	}

	result2, err := images.Process("nginx:1.14.2")
	if result2 != "backupregistry/nginx:1.14.2" {
		t.Errorf("Process(\"nginx:1.14.2\") failed, expected %v, got %v  error %v", "backupregistry/nginx:1.14.2", result2, err)
	}
}
