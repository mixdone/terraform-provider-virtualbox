package pkg

import (
	"fmt"
	"path/filepath"

	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

type LoadingType int

// now here 3 types of loading
const (
	vdiLoading LoadingType = iota
	imageloading
	empty
)

// create VM with chosen loading type
func CreateVM(vmName string, CPUs, memory int, image_path, dirName string, ltype LoadingType, vdi_size int64, os_id string, NICs [4]vbg.NIC) (*vbg.VirtualMachine, error) {
	// make path to existing vdi or create name from new vdi
	var vdiDisk string
	switch ltype {
	case vdiLoading:
		vdiDisk = image_path
	case imageloading:
		vdiDisk = filepath.Base(image_path)
		vdiDisk = vdiDisk[:len(vdiDisk)-len(filepath.Ext(vdiDisk))] + vmName + ".vdi"
		vdiDisk = filepath.Join(dirName, vdiDisk)
	}

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	storageController1 := vbg.StorageController{
		Name: "SATA Controller",
		Type: vbg.SATA,
	}

	sata := vbg.StorageControllerAttachment{
		Type:   vbg.SATA,
		Name:   "SATA Controller",
		Port:   0,
		Device: 0,
	}

	disk_VDI := vbg.Disk{
		Path:       vdiDisk,
		Format:     vbg.VDI,
		SizeMB:     vdi_size,
		Type:       "hdd",
		Controller: sata,
	}

	// Loader for the image
	storageController2 := vbg.StorageController{
		Name: "IDE Controller",
		Type: vbg.IDE,
	}

	ide := vbg.StorageControllerAttachment{
		Type:   vbg.IDE,
		Name:   "IDE Controller",
		Port:   1,
		Device: 0,
	}

	disk_ISO := vbg.Disk{
		Path:       image_path,
		Type:       "dvddrive",
		Controller: ide,
	}

	if ltype == imageloading {
		if err := vb.CreateDisk(&disk_VDI); err != nil {
			return nil, fmt.Errorf("disk creation failed: %s", err.Error())
		}
	}

	var disks []vbg.Disk
	switch ltype {
	case vdiLoading:
		disks = []vbg.Disk{disk_VDI}
	case imageloading:
		disks = []vbg.Disk{disk_VDI, disk_ISO}
	case empty:
		disks = []vbg.Disk{}
	}

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.OSType{ID: os_id},
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  disks,
		NICs:   NICs[:],
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		return nil, fmt.Errorf("VM creation failed: %s", err.Error())
	}

	if err := vb.RegisterVM(vm); err != nil {
		return nil, fmt.Errorf("failed registering vm: %s", err.Error())
	}

	// Set CPUs and memory
	if err := vb.SetCPUCount(vm, vm.Spec.CPU.Count); err != nil {
		return nil, fmt.Errorf("set CPU Count failed: %s", err.Error())
	}

	if err := vb.SetMemory(vm, vm.Spec.Memory.SizeMB); err != nil {
		return nil, fmt.Errorf("set memory failed: %s", err.Error())
	}

	if err := vb.ModifyVM(vm, []string{"network_adapter"}); err != nil {
		return nil, fmt.Errorf("set network failed: %s", err.Error())
	}

	// Connecting a disk to a virtual machine
	if ltype != empty {
		if err := vb.AddStorageController(vm, storageController1); err != nil {
			return nil, fmt.Errorf("add SATA controller error: %s", err.Error())
		}

		if err := vb.AttachStorage(vm, &disk_VDI); err != nil {
			return nil, fmt.Errorf("attach error: %s", err.Error())
		}
	}

	if ltype == imageloading {
		// Connecting the installation disk image
		if err := vb.AddStorageController(vm, storageController2); err != nil {
			return nil, fmt.Errorf("add IDE controller error: %s", err.Error())
		}

		if err := vb.AttachStorage(vm, &disk_ISO); err != nil {
			return nil, fmt.Errorf("attach error: %s", err.Error())
		}
	}

	return vm, nil
}
