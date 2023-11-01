package createvm

import (
	"log"
	"os"
	"path/filepath"

	vbg "github.com/uruddarraju/virtualbox-go"
)

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

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	err = vb.CreateVM(vm)
	if err != nil {
		log.Fatalf("VM creation failed: %s", err.Error())
	}

	err = vb.RegisterVM(vm)
	if err != nil {
		log.Fatalf("Failed registering vm")
	}

	vb.SetCPUCount(vm, vm.Spec.CPU.Count)
	vb.SetMemory(vm, vm.Spec.Memory.SizeMB)

	return dirName, vb, vm
}
