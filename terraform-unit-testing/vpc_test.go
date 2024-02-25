package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
)

func TestVPC(t *testing.T) {
	terraformOptions := &terraform.Options{
		// The path to where your Terraform configuration files are located
		TerraformDir: "../examples/resources/main.tf",
	}

	// Clean up resources with defer after testing
	defer terraform.Destroy(t, terraformOptions)

	// Initialize and apply Terraform
	terraform.InitAndApply(t, terraformOptions)

	// Write assertions here to validate the correctness of your infrastructure code
}
