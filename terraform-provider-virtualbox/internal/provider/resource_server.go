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
		Exists: resourceVirtualBoxExists,
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
		return err
	}

	if err = setState(d, vm); err != nil {
		logrus.Fatalf("Didn't manage to set VMState: %s", err.Error())
	}

	if err = d.Set("name", vm.Spec.Name); err != nil {
		logrus.Fatalf("Didn't manage to set name: %v", err.Error())
	}

	if err = d.Set("cpus", vm.Spec.CPU); err != nil {
		logrus.Fatalf("Didn't manage to set cpus: %v", err.Error())
	}

	if err = d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		logrus.Fatalf("Didn't manage to set memory: %v", err.Error())
	}

	return err
}

func poweroffVM(d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	if _, err := vb.ControlVM(vm, "poweroff"); err != nil {
		logrus.Fatalf("Unable to poweroff VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Poweroff
	return setState(d, vm)
}

func resourceVirtualBoxUpdate(d *schema.ResourceData, m interface{}) error {
	parameters := []string{}

	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
	}

	if err = poweroffVM(d, vm, vb); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
	}

	actualName := vm.Spec.Name
	newName, ok := d.Get("name").(string)
	if !ok {
		logrus.Info("Convertion name to string failed")
	}
	if actualName != newName {
		parameters = append(parameters, "name")
		vm.Spec.Name = newName
	}

	actualMemory := vm.Spec.Memory.SizeMB
	newMemory := d.Get("memory").(int)
	if actualMemory != newMemory {
		parameters = append(parameters, "memory")
		vm.Spec.Memory.SizeMB = newMemory
	}

	actualCPUCount := vm.Spec.CPU.Count
	newCPUCount := d.Get("cpus").(int)
	if actualCPUCount != newCPUCount {
		parameters = append(parameters, "cpus")
		vm.Spec.CPU.Count = newCPUCount
	}

	err = vb.ModifyVM(vm, parameters)
	if err != nil {
		logrus.Fatalf("ModifyVM failed: %s", err.Error())
	}

	id := vm.UUIDOrName()
	vm.UUID = id
	d.SetId(vm.UUID)

	if _, err := vb.ControlVM(vm, "running"); err != nil {
		logrus.Fatalf("Unable to running VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Running
	if err = setState(d, vm); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
	}

	return resourceVirtualBoxRead(d, m)
}

func resourceVirtualBoxExists(d *schema.ResourceData, m interface{}) (bool, error) {
	vb := vbg.NewVBox(vbg.Config{})
	_, err := vb.VMInfo(d.Id())
	switch err {
	case nil:
		return true, nil
	case vbg.ErrMachineNotExist:
		return false, nil
	default:
		return false, fmt.Errorf("VMInfo failed: %s", err)
	}
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
	return err
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
