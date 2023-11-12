package createvm

import (
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	vbg "github.com/uruddarraju/virtualbox-go"
)

func CreateVM(vmName string, CPUs int, memory int) (string, *vbg.VBox, *vbg.VirtualMachine) {
	dirName, err := os.MkdirTemp("./", "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	// HDD
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

	disk1 := vbg.Disk{
		Path:       filepath.Join(dirName, "disk1.vdi"),
		Format:     vbg.VDI,
		SizeMB:     10000,
		Type:       "hdd",
		Controller: sata,
	}

	// Loader for the image
	ctr1 := vbg.StorageController{
		Name: "IDE Controller",
		Type: vbg.IDE,
	}

	ide := vbg.StorageControllerAttachment{
		Type:   vbg.IDE,
		Name:   "IDE Controller",
		Port:   1,
		Device: 0,
	}

	disk2 := vbg.Disk{
		Path:       "ubuntu-23.10.1-desktop-amd64.iso",
		Type:       "dvddrive",
		Controller: ide,
	}

	logrus.Infof("disk: %s", string(disk1.Type))
	err = vb.CreateDisk(&disk1)
	if err != nil {
		logrus.Fatalf("Disk creation failed: %s", err.Error())
	}

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  []vbg.Disk{disk1, disk2},
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	err = vb.CreateVM(vm)
	if err != nil {
		logrus.Infof("VM creation failed: %s", err.Error())
	}

	logrus.Infoln("Created VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	err = vb.RegisterVM(vm)
	if err != nil {
		logrus.Fatalf("Failed registering vm")
	}

	logrus.Infoln("Register VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// Connecting a disk to a virtual machine
	err = vb.AddStorageController(vm, ctr)
	if err != nil {
		logrus.Fatalf("Add SATA controller error: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk1)
	if err != nil {
		logrus.Fatalf("Attach error: %s", err.Error())
	}

	// Connecting the installation disk image
	err = vb.AddStorageController(vm, ctr1)
	if err != nil {
		logrus.Fatalf("Add IDE controller error: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk2)
	if err != nil {
		logrus.Fatalf("Attach error %s", err.Error())
	}

	return dirName, vb, vm
}
