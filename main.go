package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	vbg "github.com/uruddarraju/virtualbox-go"
)

func main() {
	// create VM with name(vmNmae)
	vmName, CPUs, memory := "node3", 2, 1000
	dirname, vb, vm := CreateVM(vmName, CPUs, memory)
	fmt.Println(dirname)

	// get VM info
	vm, err := vb.VMInfo(vmName)
	if err != nil {
		log.Fatalf("Get info VM failed: %s", err.Error())
	}
	fmt.Printf(" name:%s\n OSType:%s\n CPUs:%d\n memory:%d\n", vm.Spec.Name, vm.Spec.OSType.Description, vm.Spec.CPU.Count, vm.Spec.Memory.SizeMB)

	//delete VM
	vb.DeleteVM(vm)
	vb.UnRegisterVM(vm)
}

func CreateVM(vmName string, CPUs, memory int) (string, *vbg.VBox, *vbg.VirtualMachine) {
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

	spec := &vbg.VirtualMachineSpec{
		Name:   vmName,
		OSType: vbg.Linux64,
		CPU:    vbg.CPU{Count: CPUs},
		Memory: vbg.Memory{SizeMB: memory},
		Disks:  []vbg.Disk{disk1},
	}

	fmt.Println(vmName, CPUs, memory)

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

	return dirName, vb, vm
}

func GetVMInfo(name string) (machine *vbg.VirtualMachine, err error) {
	vb := vbg.NewVBox(vbg.Config{})
	return vb.VMInfo(name)
}
