package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vbg "github.com/mixdone/virtualbox-go"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"virtualbox_server": resourceVM(),
		},
		ConfigureContextFunc: configVbox,
	}
}

func configVbox(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return vbg.NewVBox(vbg.Config{}), nil
}
