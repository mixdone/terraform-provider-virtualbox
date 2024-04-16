package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vbg "github.com/mixdone/virtualbox-go"
)

func dhcp() *schema.Resource {
	return &schema.Resource{
		CreateContext: dhcpServerCreate,
		ReadContext:   dhcpServerRead,
		UpdateContext: dhcpServerUpdate,
		DeleteContext: dhcpServerDelete,
		Exists:        dhcpServerExists,

		Schema: map[string]*schema.Schema{
			"server_ip": {
				Description: "server ip",
				Type:        schema.TypeString,
				Required:    true,
			},

			"lower_ip": {
				Description: "lower bound for ip addresses",
				Type:        schema.TypeString,
				Required:    true,
			},

			"upper_ip": {
				Description: "upper bound for ip addresses",
				Type:        schema.TypeString,
				Required:    true,
			},

			"network_name": {
				Description: "the name of the network where the dhcp server will be running",
				Type:        schema.TypeString,
				Required:    true,
			},

			"network_mask": {
				Description: "netmask(like ip)",
				Type:        schema.TypeString,
				Required:    true,
			},

			"enabled": {
				Description: "enabled/disabled dhcp server",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
		}}
}

func dhcpServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var dhcp vbg.DHCPServer

	dhcp.IPAddress = d.Get("server_ip").(string)
	dhcp.LowerIPAddress = d.Get("lower_ip").(string)
	dhcp.UpperIPAddress = d.Get("upper_ip").(string)
	dhcp.NetworkMask = d.Get("network_mask").(string)
	dhcp.NetworkName = d.Get("network_name").(string)
	dhcp.Enabled = d.Get("enabled").(bool)

	vb := vbg.NewVBox(vbg.Config{})
	vb.AddDHCPServer(dhcp)

	return dhcpServerRead(ctx, d, m)
}

func dhcpServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{})
	dhcp, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		diag.Errorf("dhcpInfo failed: %s", err.Error())
	}

	if err := d.Set("server_ip", dhcp.IPAddress); err != nil {
		diag.Errorf("Didn't manage to set server ip: %s", err.Error())
	}

	if err := d.Set("lower_ip", dhcp.LowerIPAddress); err != nil {
		diag.Errorf("Didn't manage to set lower ip: %s", err.Error())
	}

	if err := d.Set("upper_ip", dhcp.UpperIPAddress); err != nil {
		diag.Errorf("Didn't manage to set upper ip: %s", err.Error())
	}

	if err := d.Set("network_mask", dhcp.NetworkMask); err != nil {
		diag.Errorf("Didn't manage to set network mask: %s", err.Error())
	}

	if err := d.Set("network_name", dhcp.NetworkName); err != nil {
		diag.Errorf("Didn't manage to set network name: %s", err.Error())
	}

	return nil
}

func dhcpServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{})
	dhcpOld, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		diag.Errorf("dhcpInfo failed: %s", err.Error())
	}

	var dhcpNew vbg.DHCPServer

	dhcpNew.IPAddress = d.Get("server_ip").(string)
	dhcpNew.LowerIPAddress = d.Get("lower_ip").(string)
	dhcpNew.UpperIPAddress = d.Get("upper_ip").(string)
	dhcpNew.NetworkMask = d.Get("network_mask").(string)
	dhcpNew.NetworkName = d.Get("network_name").(string)
	dhcpNew.Enabled = d.Get("enabled").(bool)

	var parametrs []string

	if dhcpOld.IPAddress != dhcpNew.IPAddress {
		parametrs = append(parametrs, "ip")
	}

	if dhcpOld.LowerIPAddress != dhcpNew.LowerIPAddress {
		parametrs = append(parametrs, "lowerip")
	}

	if dhcpOld.UpperIPAddress != dhcpNew.UpperIPAddress {
		parametrs = append(parametrs, "upperip")
	}

	if dhcpOld.Enabled != dhcpNew.Enabled {
		parametrs = append(parametrs, "work")
	}

	if dhcpOld.NetworkMask != dhcpNew.NetworkMask {
		parametrs = append(parametrs, "netmask")
	}

	if dhcpOld.NetworkName != dhcpNew.NetworkName {
		parametrs = append(parametrs, "netname")
	}

	if err := vb.ModifyDHCPServer(dhcpNew, parametrs); err != nil {
		diag.Errorf("Modify DHCP failed: %s", err.Error())
	}

	return dhcpServerRead(ctx, d, m)
}

func dhcpServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{})
	dhcp, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		diag.Errorf("dhcpInfo failed: %s", err.Error())
	}

	if err := vb.RemoveDHCPServer(dhcp.NetworkName); err != nil {
		diag.Errorf("removeDHCP() failed: %s", err.Error())
	}

	return nil
}

func dhcpServerExists(d *schema.ResourceData, m interface{}) (bool, error) {
	vb := vbg.NewVBox(vbg.Config{})
	_, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		return false, err
	}
	return true, err
}
