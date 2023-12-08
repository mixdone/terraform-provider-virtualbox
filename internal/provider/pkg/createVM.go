package pkg

import (
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
func CreateVM(vmName string, CPUs, memory int, image_path, dirName string, ltype LoadingType) (*vbg.VirtualMachine, error) {
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
		SizeMB:     15000,
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
			logrus.Fatalf("Disk creation failed: %s", err.Error())
			return nil, err
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
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  disks,
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		logrus.Errorf("VM creation failed: %s", err.Error())
		return nil, err
	}

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Errorf("Failed registering vm")
		return nil, err
	}

	// Set CPUs and memory
	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// Connecting a disk to a virtual machine
	if ltype != empty {
		if err := vb.AddStorageController(vm, storageController1); err != nil {
			logrus.Errorf("Add SATA controller error: %s", err.Error())
			return nil, err
		}

		if err := vb.AttachStorage(vm, &disk_VDI); err != nil {
			logrus.Errorf("Attach error: %s", err.Error())
			return nil, err
		}
	}

	if ltype == imageloading {
		// Connecting the installation disk image
		if err := vb.AddStorageController(vm, storageController2); err != nil {
			logrus.Errorf("Add IDE controller error: %s", err.Error())
			return nil, err
		}

		if err := vb.AttachStorage(vm, &disk_ISO); err != nil {
			logrus.Errorf("Attach error: %s", err.Error())
			return nil, err
		}
	}

	return vm, nil
}
