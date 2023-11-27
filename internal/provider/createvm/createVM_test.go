package createvm

import (
	"fmt"
	"testing"
)

func TestProvider(t *testing.T) {

	dirname, vb, vm := CreateVM("name", 1, 10)

	fmt.Print(dirname, vb, vm)
}
