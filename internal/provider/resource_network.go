package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vbg "github.com/mixdone/virtualbox-go"
)

func resourceHostOnly() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHostOnlyCreate,
		ReadContext:   resourceHostOnlyRead,
		UpdateContext: resourceHostOnlyUpdate,
		DeleteContext: resourceHostOnlyDelete,
		Exists:        resourceHostOnlyExists,

		Schema: map[string]*schema.Schema{
			"index": {
				Type:     schema.TypeInt,
				Required: true,
			},

			"ip": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"netmask": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceHostOnlyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "~/Desktop",
	})

	var net vbg.Network
	vb.CreateNet(&net)

	return resourceHostOnlyRead(ctx, d, m)
}

func resourceHostOnlyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "~/Desktop",
	})

	nets, err := vb.HostOnlyNetInfo()

	if err != nil {
		return diag.Errorf(err.Error())
	}
	for _, net := range nets {
		if net.GUID != d.Id() {
			continue
		}

		if errors := d.Set("index", net.DeviceName[7:]); errors != nil {
			return diag.Errorf(errors.Error())
		}
		if errors := d.Set("ipv4", net.IPNet.IP); errors != nil {
			return diag.Errorf(errors.Error())
		}
		if errors := d.Set("netmask_ipv4", net.IPNet.Mask); errors != nil {
			return diag.Errorf(errors.Error())
		}

	}
	return nil
}

func resourceHostOnlyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceHostOnlyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "~/Desktop",
	})

	net := vbg.Network{
		Name: fmt.Sprintf("vboxnet%v", d.Get("index")),
	}
	if err := vb.DeleteNet(&net); err != nil {
		return diag.Errorf(err.Error())
	}

	return resourceHostOnlyRead(ctx, d, m)
}

func resourceHostOnlyExists(d *schema.ResourceData, m interface{}) (bool, error) {

	return false, nil
}
