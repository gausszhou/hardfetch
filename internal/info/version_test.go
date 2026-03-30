package info

import "testing"

func TestVersion(t *testing.T) {
	if Version != "0.1.0" {
		t.Errorf("Version = %s; want 0.1.0", Version)
	}
}
