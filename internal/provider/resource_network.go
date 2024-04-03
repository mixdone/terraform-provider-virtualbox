package provider

import (
	"context"
	"fmt"
	"strconv"

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
				ForceNew: true,
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
		BasePath: "./",
	})

	var net vbg.Network
	vb.CreateNet(&net)

	d.SetId(net.Name)

	return resourceHostOnlyRead(ctx, d, m)
}

func resourceHostOnlyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "./",
	})

	nets, err := vb.HostOnlyNetInfo()
	if err != nil {
		d.SetId("")
		return diag.Errorf(err.Error())
	}

	id := d.Id()

	for _, net := range nets {
		if net.Name != id {
			continue
		}

		index, _ := strconv.Atoi(net.DeviceName[7:])
		if errors := d.Set("index", index); errors != nil {
			return diag.Errorf(errors.Error())
		}
		if errors := d.Set("ip", net.IPNet.IP.String()); errors != nil {
			return diag.Errorf(errors.Error())
		}
		if errors := d.Set("netmask", net.IPNet.Mask.String()); errors != nil {
			return diag.Errorf(errors.Error())
		}

		break
	}
	return nil
}

func resourceHostOnlyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return nil
}

func resourceHostOnlyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "./",
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
	vb := vbg.NewVBox(vbg.Config{
		BasePath: "~/Desktop",
	})

	nets, err := vb.HostOnlyNetInfo()
	if err != nil {
		return false, fmt.Errorf("network info failed: %s", err)
	}

	id := d.Id()
	for _, net := range nets {
		if net.GUID == id {
			return true, nil
		}
	}

	return false, nil
}
