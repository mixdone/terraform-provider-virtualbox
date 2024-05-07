package provider

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

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
				Description: "host-only network adapters are restricted to IPs in the range 192.168.56.0/21. You can tell VirtualBox to allow additional IP ranges by configuring /etc/vbox/networks.conf",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"netmask": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "255.255.255.0",
			},
		},
	}
}

func resourceHostOnlyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homeDir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	netCurr := vbg.Network{
		Name: fmt.Sprintf("vboxnet%v", d.Get("index").(int)),
		Mode: vbg.NWMode_hostonly,
	}

	if err := vb.CreateNet(&netCurr); err != nil {
		return diag.Errorf(err.Error())
	}

	d.SetId(netCurr.Name)

	if ip, ok := d.GetOk("ip"); ok {
		octetsIP := strings.Split(ip.(string), ".")
		oip0, _ := strconv.Atoi(octetsIP[0])
		oip1, _ := strconv.Atoi(octetsIP[1])
		oip2, _ := strconv.Atoi(octetsIP[2])
		oip3, _ := strconv.Atoi(octetsIP[3])
		netCurr.IPNet.IP = net.IPv4(byte(oip0), byte(oip1), byte(oip2), byte(oip3))

		octetsMask := strings.Split(d.Get("netmask").(string), ".")
		omask0, _ := strconv.Atoi(octetsMask[0])
		omask1, _ := strconv.Atoi(octetsMask[1])
		omask2, _ := strconv.Atoi(octetsMask[2])
		omask3, _ := strconv.Atoi(octetsMask[3])
		netCurr.IPMask.IP = net.IPv4(byte(omask0), byte(omask1), byte(omask2), byte(omask3))

		//netCurr.IPNet.IP = net.IP(ip.(string))
		//netCurr.IPNet.IP = net.IP(d.Get("netmask").(string))

		if err := vb.ChangeNet(&netCurr); err != nil {
			return diag.Errorf(err.Error())
		}
	}

	return resourceHostOnlyRead(ctx, d, m)
}

func resourceHostOnlyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homeDir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	nets, err := vb.HostOnlyNetInfo()
	if err != nil {
		d.SetId("")
		return diag.Errorf(err.Error())
	}

	id := d.Id()

	var necessaryNetwork *vbg.Network

	for _, i := range nets {
		if i.Name == id {
			necessaryNetwork = &i
		}
	}

	index, _ := strconv.Atoi(necessaryNetwork.DeviceName[7:])
	if errors := d.Set("index", index); errors != nil {
		return diag.Errorf(errors.Error())
	}
	if errors := d.Set("ip", necessaryNetwork.IPNet.IP.String()); errors != nil {
		return diag.Errorf(errors.Error())
	}
	if errors := d.Set("netmask", necessaryNetwork.IPMask.IP.String()); errors != nil {
		return diag.Errorf(errors.Error())
	}

	return nil
}

func resourceHostOnlyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homeDir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	nets, err := vb.HostOnlyNetInfo()
	if err != nil {
		return diag.Errorf(err.Error())
	}

	var necessaryNetwork vbg.Network
	necessaryNetwork.Mode = vbg.NWMode_hostonly

	id := d.Id()

	for _, i := range nets {
		if i.Name == id {
			necessaryNetwork = i
		}
	}

	actualIP := necessaryNetwork.IPNet.IP
	newIP := d.Get("ip").(string)
	if fmt.Sprintf("%v", actualIP) != newIP {
		necessaryNetwork.IPNet.IP = net.IP(newIP)
	}

	actualIPMask := necessaryNetwork.IPMask.IP
	newIPMask := d.Get("netmask").(string)
	if actualIPMask.String() != newIPMask {
		necessaryNetwork.IPMask.IP = net.IP(newIPMask)
	}

	vb.ChangeNet(&necessaryNetwork)

	return resourceHostOnlyRead(ctx, d, m)
}

func resourceHostOnlyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homeDir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	netCurr := vbg.Network{
		Name: fmt.Sprintf("vboxnet%v", d.Get("index")),
		Mode: vbg.NWMode_hostonly,
	}
	if err := vb.DeleteNet(&netCurr); err != nil {
		return diag.Errorf(err.Error())
	}

	return nil
}

func resourceHostOnlyExists(d *schema.ResourceData, m interface{}) (bool, error) {
	homeDir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	nets, err := vb.HostOnlyNetInfo()
	if err != nil {
		return false, fmt.Errorf("network info failed: %s", err)
	}

	id := d.Id()
	for _, netCurr := range nets {
		if netCurr.Name == id {
			return true, nil
		}
	}

	return false, nil
}
