package provider

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mixdone/terraform-provider-virtualbox/internal/provider/pkg"
	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualBoxCreate,
		ReadContext:   resourceVirtualBoxRead,
		UpdateContext: resourceVirtualBoxUpdate,
		DeleteContext: resourceVirtualBoxDelete,

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

func resourceVirtualBoxCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cpus := d.Get("cpus").(int)
	memory := d.Get("memory").(int)

	homedir, _ := os.UserHomeDir()
	machinesDir := filepath.Join(homedir, "VirtualMachines")
	installedData := filepath.Join(homedir, "InstalledData")

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		logrus.Fatalf("Creation VirtualMachines foldier failed: %s", err.Error())
		return diag.Errorf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		logrus.Fatalf("Creation InstalledData foldier failed: %s", err.Error())
		return diag.Errorf("Creation InstalledData foldier failed: %s", err.Error())
	}

	logrus.Info("1")

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
				return diag.Errorf("File dowload failed: %s", err.Error())
			}

			if filepath.Ext(filepath.Base(filename)) != ".iso" {
				imagePath, err := pkg.UnpackImage(filename, installedData)
				if err != nil {
					logrus.Fatalf("File unpaking failed: %s", err.Error())
					return diag.Errorf("File unpaking failed: %s", err.Error())
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

	logrus.Info("2")

	vm, err := pkg.CreateVM(name, cpus, memory, image, machinesDir, ltype)
	if err != nil {
		logrus.Fatalf("Creation VM failed: %s", err.Error())
		return diag.Errorf("Creation VM failed: %s", err.Error())
	}

	d.SetId(vm.UUID)
	return resourceVirtualBoxRead(ctx, d, m)
}

func resourceVirtualBoxRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logrus.Info("3")
	vb := vbg.NewVBox(vbg.Config{})
	logrus.Info("4")
	vm, err := vb.VMInfo(d.Get("name").(string))

	if err != nil {
		d.SetId("")
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}
	logrus.Info("5")
	if err := setState(d, vm); err != nil {
		logrus.Fatalf("Didn't manage to set VMState: %s", err.Error())
		return diag.Errorf("Didn't manage to set VMState: %s", err.Error())
	}

	if err := d.Set("name", vm.Spec.Name); err != nil {
		logrus.Fatalf("Didn't manage to set name: %s", err.Error())
		return diag.Errorf("Didn't manage to set name: %s", err.Error())
	}

	if err := d.Set("cpus", vm.Spec.CPU); err != nil {
		logrus.Fatalf("Didn't manage to set cpus: %s", err.Error())
		return diag.Errorf("Didn't manage to set cpus: %s", err.Error())
	}

	if err := d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		logrus.Fatalf("Didn't manage to set memory: %s", err.Error())
		return diag.Errorf("Didn't manage to set memory: %s", err.Error())
	}

	return nil
}

func poweroffVM(ctx context.Context, d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
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

func resourceVirtualBoxUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	if err := poweroffVM(ctx, d, vm, vb); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return diag.Errorf("Setting state failed: %s", err.Error())
	}

	actualName := vm.Spec.Name
	newName := d.Get("name").(string)
	if actualName != newName {
		vm.Spec.Name = newName
	}

	actualMemory := vm.Spec.Memory.SizeMB
	newMemory := d.Get("memory").(int)
	if actualMemory != newMemory {
		if err = vb.SetMemory(vm, d.Get("memory").(int)); err != nil {
			logrus.Fatalf("Setting memory faild: %s", err.Error())
			return diag.Errorf("Setting memory faild: %s", err.Error())
		}
		vm.Spec.Memory.SizeMB = newMemory
	}

	actualCPUCount := vm.Spec.CPU.Count
	newCPUCount := d.Get("cpus").(int)
	if actualCPUCount != newCPUCount {
		if err = vb.SetCPUCount(vm, d.Get("cpus").(int)); err != nil {
			logrus.Fatalf("Setting CPUs faild: %s", err.Error())
			return diag.Errorf("Setting CPUs faild: %s", err.Error())
		}
		vm.Spec.CPU.Count = newCPUCount
	}

	id := vm.UUIDOrName()
	vm.UUID = id
	d.SetId(vm.UUID)

	if _, err = vb.Start(vm); err != nil {
		logrus.Fatalf("Unable to running VM: %s", err.Error())
		return diag.Errorf("Unable to running VM: %s", err.Error())
	}

	vm.Spec.State = vbg.Running
	if err = setState(d, vm); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return diag.Errorf("Setting state failed: %s", err.Error())
	}

	return resourceVirtualBoxRead(ctx, d, m)
}

func resourceVirtualBoxDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	vb := vbg.NewVBox(vbg.Config{})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	if err = vb.UnRegisterVM(vm); err != nil {
		logrus.Fatalf("VM Unregiste failed: %s", err.Error())
		return diag.Errorf("VM Unregiste failed: %s", err.Error())
	}

	if err = vb.DeleteVM(vm); err != nil {
		logrus.Fatalf("VM deletion failed: %s", err.Error())
		return diag.Errorf("VM deletion failed: %s", err.Error())
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
