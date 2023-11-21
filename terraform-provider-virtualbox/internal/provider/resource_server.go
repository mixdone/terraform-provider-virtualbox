package provider

import (
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/mixdone/terraform-provider-virtualbox/internal/provider/pkg"
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
				Default:  "running",
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
	name := d.Get("name").(string)
	cpus := d.Get("cpus").(int)
	memory := d.Get("memory").(int)

	homedir, _ := os.UserHomeDir()
	machinesDir := filepath.Join(homedir, "VirtualMachines")
	installedData := filepath.Join(homedir, "InstalledData")

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		logrus.Fatalf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		logrus.Fatalf("Creation InstalledData foldier failed: %s", err.Error())
	}

	var ltype pkg.LoadingType

	im, ok := d.GetOk("image")
	image := im.(string)
	if !ok {
		url, ok := d.GetOk("url")
		if !ok {
			ltype = 2
		} else {
			filename, err := pkg.FileDownload(url.(string), homedir)
			if err != nil {
				logrus.Fatalf("File dowload failed: %s", err.Error())
				return err
			}

			if filepath.Ext(filepath.Base(filename)) != ".iso" {
				imagePath, err := pkg.UnpackImage(filename, installedData)
				if err != nil {
					logrus.Fatalf("File unpacking failed")
				}
				image = imagePath
			} else {
				image = filename
			}
		}
	} else {
		if filepath.Ext(filepath.Base(image)) == ".iso" {
			ltype = 1
		} else {
			ltype = 0
		}
	}

	vm, err := pkg.CreateVM(name, cpus, memory, image, machinesDir, ltype)
	if err != nil {
		logrus.Fatalf("Creation failde: %s", err.Error())
		return err
	}

	d.SetId(vm.UUID)
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
		return err
	}

	if err = d.Set("name", vm.Spec.Name); err != nil {
		logrus.Fatalf("Didn't manage to set name: %v", err.Error())
		return err
	}

	if err = d.Set("cpus", vm.Spec.CPU); err != nil {
		logrus.Fatalf("Didn't manage to set cpus: %v", err.Error())
		return err
	}

	if err = d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		logrus.Fatalf("Didn't manage to set memory: %v", err.Error())
		return err
	}

	return nil
}

func poweroffVM(d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	if _, err := vb.Stop(vm); err != nil {
		logrus.Fatalf("Unable to poweroff VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Poweroff
	return setState(d, vm)
}

func resourceVirtualBoxUpdate(d *schema.ResourceData, m interface{}) error {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return err
	}

	if err = poweroffVM(d, vm, vb); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return err
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
		if err = vb.SetMemory(vm, d.Get("memory").(int)); err != nil {
			logrus.Fatalf("Setting memory faild: %s", err.Error())
			return err
		}
		vm.Spec.Memory.SizeMB = newMemory
	}

	actualCPUCount := vm.Spec.CPU.Count
	newCPUCount := d.Get("cpus").(int)
	if actualCPUCount != newCPUCount {
		if err = vb.SetCPUCount(vm, d.Get("cpus").(int)); err != nil {
			logrus.Fatalf("Setting CPUs faild: %s", err.Error())
			return err
		}
		vm.Spec.CPU.Count = newCPUCount
	}

	id := vm.UUIDOrName()
	vm.UUID = id
	d.SetId(vm.UUID)

	if _, err = vb.Start(vm); err != nil {
		logrus.Fatalf("Unable to running VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Running
	if err = setState(d, vm); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return err
	}

	return resourceVirtualBoxRead(d, m)
}

func resourceVirtualBoxDelete(d *schema.ResourceData, m interface{}) error {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return err
	}

	if err = vb.UnRegisterVM(vm); err != nil {
		logrus.Fatalf("VM Unregiste failed: %s", err.Error())
		return err
	}

	if err = vb.DeleteVM(vm); err != nil {
		logrus.Fatalf("VM deletion failed: %s", err.Error())
		return err
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
