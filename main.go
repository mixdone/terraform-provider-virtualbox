package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	vbg "github.com/uruddarraju/virtualbox-go"
)

func main() {
	vmName := "node1"
	dirname, vb := CreateVM(vmName)
	fmt.Println(dirname)
	vm, err := GetVMInfo(vmName)
	if err != nil {
		log.Fatalf("Get info VM failed: %s", err.Error())
	}
	defer vb.DeleteVM(vm)
	fmt.Printf(" name:%s\n OSType:%s\n CPUs:%d\n memory:%d\n", vm.Spec.Name, vm.Spec.OSType.Description, vm.Spec.CPU.Count, vm.Spec.Memory.SizeMB)
}

func CreateVM(vmName string) (string, *vbg.VBox) {
	dirName, err := os.MkdirTemp("", "vbm")
	if err != nil {
		log.Fatalf("Tempdir creation failed: %s", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	disk1 := vbg.Disk{
		Path:   filepath.Join(dirName, "disk1.vdi"),
		Format: vbg.VDI,
		SizeMB: 10,
	}

	err = vb.CreateDisk(&disk1)
	if err != nil {
		log.Fatalf("Disk creation failed: %s", err.Error())
	}

	vm := &vbg.VirtualMachine{}
	vm.Spec.Name = vmName
	vm.Spec.OSType = vbg.Linux64
	vm.Spec.CPU.Count = 2
	vm.Spec.Memory.SizeMB = 1000
	vm.Spec.Disks = []vbg.Disk{disk1}

	err = vb.CreateVM(vm)
	if err != nil {
		log.Fatalf("VM creation failed: %s", err.Error())
	}

	err = vb.RegisterVM(vm)
	if err != nil {
		log.Fatalf("Failed registering vm")
	}

	return dirName, vb
}

func GetVMInfo(name string) (machine *vbg.VirtualMachine, err error) {
	vb := vbg.NewVBox(vbg.Config{})
	return vb.VMInfo(name)
}
