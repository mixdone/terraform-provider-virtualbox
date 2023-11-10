package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	vm "github.com/mixdone/terraform-provider-virtualbox/internal/provider/createvm"
	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
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
				Default:  "poweroff",
			},
			"image": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},
	}
}

func resourceVirtualBoxCreate(d *schema.ResourceData, m interface{}) error {
	name, ok := d.Get("name").(string)
	if !ok {
		logrus.Info("Convertion name to string failed")
	}
	cpus := d.Get("cpus").(int)
	memory := d.Get("memory").(int)
	dirname, vb, vm := vm.CreateVM(name, cpus, memory)

	d.SetId(vm.UUID)

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

	if err = setState(d, vm); err != nil {
		logrus.Fatalf("Didn't manage to set VMState: %s", err.Error())
	}

	err = d.Set("name", vm.Spec.Name)
	if err != nil {
		logrus.Fatalf("Didn't manage to set name: %v", err.Error())
	}
	err = d.Set("cpus", vm.Spec.CPU)
	if err != nil {
		logrus.Fatalf("Didn't manage to set cpus: %v", err.Error())
	}
	err = d.Set("memory", vm.Spec.Memory.SizeMB)
	if err != nil {
		logrus.Fatalf("Didn't manage to set memory: %v", err.Error())
	}

	return nil
}

func poweroffVM(d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	_, err := vb.Stop(vm)
	if err != nil {
		logrus.Fatalf("Unable to poweroff VM: %s", err.Error())
	}

	vm.Spec.State = vbg.Poweroff
	err = setState(d, vm)
	return err
}

func resourceVirtualBoxUpdate(d *schema.ResourceData, m interface{}) error {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())
	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
	}

	err = poweroffVM(d, vm, vb)
	if err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
	}

	actualName := vm.Spec.Name
	newName, ok := d.Get("name").(string)
	if !ok {
		logrus.Info("Convertion name to string failed")
	}
	if actualName != newName {
		vm.Spec.Name = newName
	}

	actualMemory := vm.Spec.Memory.SizeMB
	newMemory := d.Get("memory").(int)
	if actualMemory != newMemory {
		err = vb.SetMemory(vm, d.Get("memory").(int))
		if err != nil {
			logrus.Fatalf("Setting memory faild: %s", err.Error())
		}
		vm.Spec.Memory.SizeMB = newMemory
	}

	actualCPUCount := vm.Spec.CPU.Count
	newCPUCount := d.Get("cpus").(int)
	if actualCPUCount != newCPUCount {
		err = vb.SetCPUCount(vm, d.Get("cpus").(int))
		if err != nil {
			logrus.Fatalf("Setting CPUs faild: %s", err.Error())
		}
		vm.Spec.CPU.Count = newCPUCount
	}

	id := vm.UUIDOrName()
	vm.UUID = id
	d.SetId(vm.UUID)

	_, err = vb.Start(vm)
	if err != nil {
		logrus.Fatalf("Unable to running VM: %s", err.Error())
	}

	vm.Spec.State = vbg.Running
	err = setState(d, vm)
	if err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
	}

	return resourceVirtualBoxRead(d, m)
}

func resourceVirtualBoxDelete(d *schema.ResourceData, m interface{}) error {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
	}

	if err = vb.UnRegisterVM(vm); err != nil {
		logrus.Fatalf("VM Unregiste failed: %s", err.Error())
	}

	if err = vb.DeleteVM(vm); err != nil {
		logrus.Fatalf("VM deletion failed: %s", err.Error())
	}
	return nil
}

func setState(d *schema.ResourceData, vm *vbg.VirtualMachine) error {
	var err error
	switch vm.Spec.State {
	case vbg.Poweroff:
		err = d.Set("status", "poweroff")
	case vbg.Running:
		err = d.Set("status", "running")
	case vbg.Paused:
		err = d.Set("status", "paused")
	case vbg.Saved:
		err = d.Set("status", "saved")
	case vbg.Aborted:
		err = d.Set("status", "aborted")
	}
	return err
}
