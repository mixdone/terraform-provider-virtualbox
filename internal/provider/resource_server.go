package provider

import (
	"context"
	"fmt"
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
		Exists:        resourceVirtualBoxExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"basedir": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "VMs",
				ForceNew: true,
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

func resourceVirtualBoxCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	cpus := d.Get("cpus").(int)
	memory := d.Get("memory").(int)

	homedir, _ := os.UserHomeDir()
	machinesDir := filepath.Join(homedir, d.Get("basedir").(string))
	installedData := filepath.Join(homedir, "InstalledData")

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		logrus.Fatalf("Creation VirtualMachines foldier failed: %s", err.Error())
		return diag.Errorf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		logrus.Fatalf("Creation InstalledData foldier failed: %s", err.Error())
		return diag.Errorf("Creation InstalledData foldier failed: %s", err.Error())
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

	vm, err := pkg.CreateVM(name, cpus, memory, image, machinesDir, ltype)
	if err != nil {
		logrus.Fatalf("Creation VM failed: %s", err.Error())
		return diag.Errorf("Creation VM failed: %s", err.Error())
	}

	d.SetId(vm.UUIDOrName())
	return resourceVirtualBoxRead(ctx, d, m)
}

func resourceVirtualBoxRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		d.SetId("")
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	if err := setState(d, vm); err != nil {
		logrus.Fatalf("Didn't manage to set VMState: %s", err.Error())
		return diag.Errorf("Didn't manage to set VMState: %s", err.Error())
	}

	if err := d.Set("name", vm.Spec.Name); err != nil {
		logrus.Fatalf("Didn't manage to set name: %s", err.Error())
		return diag.Errorf("Didn't manage to set name: %s", err.Error())
	}

	if err := d.Set("cpus", vm.Spec.CPU.Count); err != nil {
		logrus.Fatalf("Didn't manage to set cpus: %s", err.Error())
		return diag.Errorf("Didn't manage to set cpus: %s", err.Error())
	}

	if err := d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		logrus.Fatalf("Didn't manage to set memory: %s", err.Error())
		return diag.Errorf("Didn't manage to set memory: %s", err.Error())
	}

	if err := d.Set("basedir", d.Get("basedir").(string)); err != nil {
		logrus.Fatalf("Didn't manage to set basedir: %s", err.Error())
		return diag.Errorf("Didn't manage to set basedir: %s", err.Error())
	}

	return nil
}

func poweroffVM(ctx context.Context, d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	if _, err := vb.ControlVM(vm, "poweroff"); err != nil {
		logrus.Fatalf("Unable to poweroff VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Poweroff
	return nil
}

func resourceVirtualBoxUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})
	vm, err := vb.VMInfo(d.Id())

	parameters := []string{}

	if err != nil {
		logrus.Fatalf("VMInfo failed: %s", err.Error())
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	if err = poweroffVM(ctx, d, vm, vb); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return diag.Errorf("Setting state failed: %s", err.Error())
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

	if len(parameters) != 0 {
		err = vb.ModifyVM(vm, parameters)
		if err != nil {
			logrus.Fatalf("ModifyVM failed: %s", err.Error())
			return diag.Errorf("ModifyVM failed: %s", err.Error())
		}
	}

	vm.UUID = vm.UUIDOrName()
	d.SetId(vm.UUIDOrName())

	status := d.Get("status").(string)
	logrus.Printf("%s -> %s", vm.Spec.State, status)
	if status != string(vm.Spec.State) {
		if _, err := vb.ControlVM(vm, status); err != nil {
			logrus.Fatalf("Unable to running VM: %s", err.Error())
			return diag.Errorf("Unable to running VM: %s", err.Error())
		}
		logrus.Printf("%s -> %s", vm.Spec.State, status)
		vm.Spec.State = vbg.VirtualMachineState(status)
		if err = setState(d, vm); err != nil {
			logrus.Fatalf("Setting state failed: %s", err.Error())
			return diag.Errorf("Setting state failed: %s", err.Error())
		}
	}

	return resourceVirtualBoxRead(ctx, d, m)
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

func resourceVirtualBoxDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})
	vm, err := vb.VMInfo(d.Id())

	if err = poweroffVM(ctx, d, vm, vb); err != nil {
		logrus.Fatalf("Setting state failed: %s", err.Error())
		return diag.Errorf("Setting state failed: %s", err.Error())
	}

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

	machineDir := filepath.Join(homedir, d.Get("basedir").(string))
	if err := os.RemoveAll(machineDir); err != nil {
		logrus.Fatalf("Can't clear the data: %s", err.Error())
		return diag.Errorf("Can't clear the data: %s", err.Error())
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
