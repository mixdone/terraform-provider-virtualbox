package main

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	vbg "github.com/mixdone/virtualbox-go"
)

func main() {
	// create VM with name(vmNmae)
	vmName, CPUs, memory := "Test", 1, 1024
	CreateVM(vmName, CPUs, memory)
}

func CreateVM(vmName string, CPUs, memory int) (string, *vbg.VBox, *vbg.VirtualMachine) {
	dirName, err := os.MkdirTemp("./", "VirtualBox VMs")
	if err != nil {
		log.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	// Жесткий диск
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

	// Загрузчик для образа
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

	log.Infof("disk: %s", string(disk1.Type))
	err = vb.CreateDisk(&disk1)
	if err != nil {
		log.Fatalf("Disk creation failed: %s", err.Error())
	}

	// Параметры вируальной машины
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

	fmt.Println("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	err = vb.CreateVM(vm)
	if err != nil {
		log.Fatalf("VM creation failed: %s", err.Error())
	}

	fmt.Println("Created VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	err = vb.RegisterVM(vm)
	if err != nil {
		log.Fatalf("Failed registering vm")
	}

	fmt.Println("Register VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// подключение диска к виртуальной машине
	err = vb.AddStorageController(vm, ctr)
	if err != nil {
		log.Fatalf("ctr: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk1)
	if err != nil {
		log.Fatalf("strorage: %s", err.Error())
	}

	// Подключаем установочный образ диска
	err = vb.AddStorageController(vm, ctr1)
	if err != nil {
		log.Fatalf("ctr1: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk2)
	if err != nil {
		log.Fatalf("storage %s", err.Error())
	}

	return dirName, vb, vm
}

func GetVMInfo(name string) (machine *vbg.VirtualMachine, err error) {
	vb := vbg.NewVBox(vbg.Config{})
	return vb.VMInfo(name)
}
