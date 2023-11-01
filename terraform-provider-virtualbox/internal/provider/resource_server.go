package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mixdone/terraform-provider-virtualbox/internal/provider/createvm"
	"github.com/sirupsen/logrus"
	vbg "github.com/uruddarraju/virtualbox-go"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		Create: resourceVirtualBoxCreate,
		Read:   resourceVirtualBoxRead,
		Update: resourceVirtualBoxUpdate,
		Delete: resourceVirtualBoxDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  128,
			},

			"cpus": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "running",
			},
			"image": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVirtualBoxCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	cpus := d.Get("cpus").(int)
	memory := d.Get("memory").(int)
	dirname, vb, vm := createvm.CreateVM(name, cpus, memory)

	fmt.Print(dirname, vb, vm)
	return resourceVirtualBoxRead(d, m)
}

func resourceVirtualBoxRead(d *schema.ResourceData, m interface{}) error {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())
	if err != nil {
		d.SetId("")
		return nil
	}

	err = d.Set("name", vm.Spec.Name)
	if err != nil {
		logrus.Fatalf("can't set name: %v", err.Error())
	}
	err = d.Set("CPUs", vm.Spec.CPU)
	if err != nil {
		logrus.Fatalf("can't set cpus: %v", err.Error())
	}
	err = d.Set("memory", vm.Spec.Memory.SizeMB)
	if err != nil {
		logrus.Fatalf("can't set memory: %v", err.Error())
	}

	return nil
}

func resourceVirtualBoxUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceVirtualBoxRead(d, m)
}

func resourceVirtualBoxDelete(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
