package provider_test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

func TestVirtualMachineCreation(t *testing.T) {
	t.Parallel()

	terraformOptions := &terraform.Options{
		// The path to where your Terraform configuration files are located
		TerraformDir: "../examples/resources",
		/*Vars: map[string]interface{}{
			"count":  0,
			"cpus":   3,
			"memory": 1000,
			"status": "running",
			"os_id":  "Windows7_64",
		},*/
	}

	// Clean up resources with defer after testing
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	vmName := terraform.Output(t, terraformOptions, "name")
	vmBasedir := terraform.Output(t, terraformOptions, "basedir")
	vmCPUs := terraform.Output(t, terraformOptions, "cpus")
	vmMemory := terraform.Output(t, terraformOptions, "memory")
	vmStatus := terraform.Output(t, terraformOptions, "status")

	expName := "VM_without_image-01"
	expDir := "VM_without_image-01"

	assert.Equal(t, expName, vmName)
	assert.Equal(t, expDir, vmBasedir)
	assert.Equal(t, "3", vmCPUs)
	assert.Equal(t, "1000", vmMemory)
	assert.Equal(t, "poweroff", vmStatus)
}
