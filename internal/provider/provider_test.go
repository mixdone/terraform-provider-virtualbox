package provider

import (
	"testing"
)

func TestProviderInstantiation(t *testing.T) {
	s := Provider()

	if s == nil {
		t.Fatalf("Cannot instantiate Provider")
	}
}
