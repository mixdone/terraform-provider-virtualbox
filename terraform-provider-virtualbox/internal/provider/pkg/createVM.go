package pkg

import (
	"os"
	"path/filepath"

	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

func CreateVM_using_VDI(vmName string, CPUs int, memory int, dirName, diskVDI string) (*vbg.VirtualMachine, error) {
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	ctr := vbg.StorageController{
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
		Path:       filepath.Join(dirName, diskVDI),
		Format:     vbg.VDI,
		SizeMB:     32000,
		Type:       "hdd",
		Controller: sata,
	}

	if err := vb.CreateDisk(&disk_VDI); err != nil {
		logrus.Fatalf("Disk creation failed: %s", err.Error())
		return nil, err
	}

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  []vbg.Disk{disk_VDI},
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		logrus.Infof("VM creation failed: %s", err.Error())
		return nil, err
	}

	logrus.Infoln("Created VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Fatalf("Failed registering vm")
		return nil, err
	}

	logrus.Infoln("Register VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// Connecting a disk to a virtual machine
	if err := vb.AddStorageController(vm, ctr); err != nil {
		logrus.Fatalf("Add SATA controller failed: %s", err.Error())
		return nil, err
	}

	if err := vb.AttachStorage(vm, &disk_VDI); err != nil {
		logrus.Fatalf("Attach storage failed: %s", err.Error())
		return nil, err
	}
	return vm, nil
}

func CreateVM_using_ISO(vmName string, CPUs int, memory int, path_to_iso_file, dirName string) (*vbg.VirtualMachine, error) {
	iso_name := filepath.Base(path_to_iso_file)
	iso_name = iso_name[:len(iso_name)-len(filepath.Ext(iso_name))] + ".vdi"

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
		Path:       filepath.Join(dirName, iso_name),
		Format:     vbg.VDI,
		SizeMB:     10000,
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
		Path:       path_to_iso_file,
		Type:       "dvddrive",
		Controller: ide,
	}

	if err := vb.CreateDisk(&disk_VDI); err != nil {
		logrus.Fatalf("Disk creation failed: %s", err.Error())
		return nil, err
	}

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  []vbg.Disk{disk_VDI, disk_ISO},
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		logrus.Fatalf("VM creation failed: %s", err.Error())
		return nil, err
	}

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Fatalf("Failed registering vm")
		return nil, err
	}

	// Set CPUs and memory
	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// Connecting a disk to a virtual machine
	if err := vb.AddStorageController(vm, storageController1); err != nil {
		logrus.Fatalf("Add SATA controller error: %s", err.Error())
		return nil, err
	}

	if err := vb.AttachStorage(vm, &disk_VDI); err != nil {
		logrus.Fatalf("Attach error: %s", err.Error())
		return nil, err
	}

	// Connecting the installation disk image
	if err := vb.AddStorageController(vm, storageController2); err != nil {
		logrus.Fatalf("Add IDE controller error: %s", err.Error())
		return nil, err
	}

	if err := vb.AttachStorage(vm, &disk_ISO); err != nil {
		logrus.Fatalf("Attach error %s", err.Error())
		return nil, err
	}

	return vm, nil
}

func CreateVM(vmName string, CPUs int, memory int) (string, *vbg.VBox, *vbg.VirtualMachine) {
	dirName, err := os.MkdirTemp("./", "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  []vbg.Disk{},
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		logrus.Infof("VM creation failed: %s", err.Error())
	}

	logrus.Infoln("Created VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Fatalf("Failed registering vm")
	}

	logrus.Infoln("Register VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	return dirName, vb, vm
}
