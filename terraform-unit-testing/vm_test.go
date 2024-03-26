package provider_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestVirtualMachineCreation2(t *testing.T) {
	t.Parallel()
	terraformOptions := &terraform.Options{
		TerraformDir: "../examples/resources",
	}

	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	// Checking the attributes of a virtual machine
	vmName := terraform.Output(t, terraformOptions, "name_3")
	vmCPUs := terraform.Output(t, terraformOptions, "cpus_3")
	vmMemory := terraform.Output(t, terraformOptions, "memory_3")
	// vmURL := terraform.Output(t, terraformOptions, "url")
	vmStatus := terraform.Output(t, terraformOptions, "status_3")
	vmVDISize := terraform.Output(t, terraformOptions, "vdi_size_3")

	assert.Equal(t, "VM_VDI-01", vmName)
	assert.Equal(t, "2", vmCPUs)
	assert.Equal(t, "500", vmMemory)
	// assert.Equal(t, "github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz", vmURL)
	assert.Equal(t, "poweroff", vmStatus)
	assert.Equal(t, "25000", vmVDISize)
}
