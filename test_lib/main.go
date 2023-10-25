package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	vbg "github.com/uruddarraju/virtualbox-go"
)

func main() {
	name := "newVM"
	vb, vm := CreateVM(name)
	fmt.Printf("Created a machine with a name %v in the directory %v\n", name, vb.Config.BasePath)

	vm, err := vb.VMInfo(name)
	if err != nil {
		fmt.Errorf("Get info VM failed: %s", err.Error())
	}
	fmt.Printf("Information about the virtual machine:\n Name: %v\n CPUs: %v\n Memory: %v\n OSType: %v\n", vm.Spec.Name, vm.Spec.CPU.Count, vm.Spec.Memory.SizeMB, vm.Spec.OSType.ID)

	err = vb.UnRegisterVM(vm)
	if err != nil {
		fmt.Errorf("Making the VM unregistered failed: %s", err.Error())
	}
	fmt.Println("Unregistered VM")

	err = vb.DeleteVM(vm)
	if err != nil {
		fmt.Errorf("Deleting VM failed: %s", err.Error())
	}
	fmt.Println("Deleted VM")

}

func CreateVM(name string) (*vbg.VBox, *vbg.VirtualMachine) {
	dirName, err := os.MkdirTemp("", "vbm")
	if err != nil {
		fmt.Errorf("Tempdir creation failed %v", err)
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	disk1 := vbg.Disk{
		Path:   filepath.Join(dirName, "disk1.vdi"), // Path represents the absolute path in the system where the disk is stored, normally is under the vm folder
		Format: vbg.VDI,
		SizeMB: 10,
	}

	err = vb.CreateDisk(&disk1)
	if err != nil {
		fmt.Errorf("CreateDisk failed %v", err)
	}

	vm := &vbg.VirtualMachine{}
	vm.Spec.Name = name //Name identifies the vm and is also used in forming full path
	vm.Spec.OSType = vbg.Linux64
	vm.Spec.CPU.Count = 2
	vm.Spec.Memory.SizeMB = 1000
	vm.Spec.Disks = []vbg.Disk{disk1}

	err = vb.CreateVM(vm)
	if err != nil {
		log.Fatalf("Failed creating vm %v", err)
	}

	err = vb.RegisterVM(vm)
	if err != nil {
		log.Fatalf("Failed registering vm")
	}

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	return vb, vm
}

func GetVMInfo(name string) (machine *vbg.VirtualMachine, err error) {
	vb := vbg.NewVBox(vbg.Config{})
	return vb.VMInfo(name)
}
