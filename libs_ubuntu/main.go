package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	vbg "github.com/mixdone/virtualbox-go"
)

func main() {
	// create VM with name(vmNmae)
	vmName, CPUs, memory := "addNIC01", 1, 1024
	_, vb, vm, _ := CreateVM(vmName, CPUs, memory)

	netWorkAddNIC(vb, vm)

	//logrus.Infoln(vm1.Spec.Name, vm1.Spec.CPU.Count, vm1.Spec.Memory.SizeMB)
	//vb.UnRegisterVM(vm)
	//vb.DeleteVM(vm)
}

func netWorkAddNIC(vb *vbg.VBox, vm *vbg.VirtualMachine) {

	nic := vbg.NIC{
		// necessary fields
		Index:       3,
		NetworkName: "name",
		Mode:        vbg.NWMode_hostonly,
		Type:        vbg.NIC_82540EM,
	}

	if err := vb.AddNic(vm, &nic); err != nil {
		logrus.Fatalf("Add nic error: %s", err.Error())
	}
}

func CreateVM(vmName string, CPUs, memory int) (string, *vbg.VBox, *vbg.VirtualMachine, vbg.Config) {
	dirName, err := os.MkdirTemp("./", "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	config := vbg.Config{
		BasePath: dirName,
	}

	vb := vbg.NewVBox(config)

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
		Path:       filepath.Join(dirName, "disk.vdi"),
		Format:     vbg.VDI,
		SizeMB:     20000,
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

	logrus.Infof("disk: %s", string(disk1.Type))
	err = vb.CreateDisk(&disk1)
	if err != nil {
		logrus.Fatalf("Disk creation failed: %s", err.Error())
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
		logrus.Fatalf("VM creation failed: %s", err.Error())
	}

	fmt.Println("Created VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	err = vb.RegisterVM(vm)
	if err != nil {
		logrus.Fatalf("Failed registering vm")
	}

	fmt.Println("Register VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	// подключение диска к виртуальной машине
	err = vb.AddStorageController(vm, ctr)
	if err != nil {
		logrus.Fatalf("ctr: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk1)
	if err != nil {
		logrus.Fatalf("strorage: %s", err.Error())
	}

	// Подключаем установочный образ диска
	err = vb.AddStorageController(vm, ctr1)
	if err != nil {
		logrus.Fatalf("ctr1: %s", err.Error())
	}

	err = vb.AttachStorage(vm, &disk2)
	if err != nil {
		logrus.Fatalf("storage %s", err.Error())
	}

	return dirName, vb, vm, config
}

func GetVMInfo(name string, config vbg.Config) (*vbg.VirtualMachine, error) {
	vb := vbg.NewVBox(config)
	vm, err := vb.VMInfo(name)
	return vm, err
}
