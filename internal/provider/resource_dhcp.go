package provider

import (
	"context"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

// resourceDHCP returns schema for DHCP resource.
// it defines structure of DHCP resource including its attributes and CRUD operations.
func resourceDHCP() *schema.Resource {
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
				ForceNew:    true,
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

// dhcpServerCreate creates new DHCP server.
// it retrieves DHCP configuration parameters from resource data, creates DHCP server
// using VirtualBox API and stores DHCP server's ID.
func dhcpServerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	dhcp := vbg.DHCPServer{}

	dhcp.IPAddress = d.Get("server_ip").(string)
	dhcp.LowerIPAddress = d.Get("lower_ip").(string)
	dhcp.UpperIPAddress = d.Get("upper_ip").(string)
	dhcp.NetworkMask = d.Get("network_mask").(string)
	dhcp.NetworkName = d.Get("network_name").(string)
	dhcp.Enabled = d.Get("enabled").(bool)

	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: homedir})
	if _, err := vb.AddDHCPServer(dhcp); err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return diag.Errorf("add dhcpserver failed: %s", err.Error())
		}
	}
	d.SetId(dhcp.NetworkName)

	return dhcpServerRead(ctx, d, m)
}

// dhcpServerRead reads DHCP server configuration.
// it retrieves DHCP configuration parameters from VirtualBox API and sets them in resource data.
func dhcpServerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: homedir})
	dhcp, err := vb.DHCPInfo(d.Get("network_name").(string))
	logrus.Info("hello4")
	if err != nil {
		return diag.Errorf("dhcpInfo failed: %s", err.Error())
	}

	if err := d.Set("server_ip", dhcp.IPAddress); err != nil {
		return diag.Errorf("Didn't manage to set server ip: %s", err.Error())
	}

	if err := d.Set("lower_ip", dhcp.LowerIPAddress); err != nil {
		return diag.Errorf("Didn't manage to set lower ip: %s", err.Error())
	}

	if err := d.Set("upper_ip", dhcp.UpperIPAddress); err != nil {
		return diag.Errorf("Didn't manage to set upper ip: %s", err.Error())
	}

	if err := d.Set("network_mask", dhcp.NetworkMask); err != nil {
		return diag.Errorf("Didn't manage to set network mask: %s", err.Error())
	}

	if err := d.Set("network_name", dhcp.NetworkName); err != nil {
		return diag.Errorf("Didn't manage to set network name: %s", err.Error())
	}

	if err := d.Set("enabled", dhcp.Enabled); err != nil {
		return diag.Errorf("Didn't manage to set state: %s", err.Error())
	}

	return nil
}

// dhcpServerUpdate updates DHCP server configuration.
// it retrieves both old and new DHCP configurations, compares them, and modifies DHCP server.
func dhcpServerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: homedir})
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

	if err := vb.ModifyDHCPServer(dhcpNew, parametrs); err != nil {
		diag.Errorf("Modify DHCP failed: %s", err.Error())
	}

	return dhcpServerRead(ctx, d, m)
}

// dhcpServerDelete deletes DHCP server.
// it retrieves DHCP server configuration and removes it using VirtualBox API.
func dhcpServerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: homedir})
	dhcp, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		diag.Errorf("dhcpInfo failed: %s", err.Error())
	}

	if err := vb.RemoveDHCPServer(dhcp.NetworkName); err != nil {
		diag.Errorf("removeDHCP() failed: %s", err.Error())
	}

	return nil
}

// dhcpServerExists checks if DHCP server exists.
// it verifies existence of DHCP server configuration.
func dhcpServerExists(d *schema.ResourceData, m interface{}) (bool, error) {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: homedir})
	_, err := vb.DHCPInfo(d.Get("network_name").(string))
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			return false, err
		} else {
			return true, nil
		}
	}
	return true, err
}
