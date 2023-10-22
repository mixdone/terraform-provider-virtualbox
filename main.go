package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	vbg "github.com/uruddarraju/virtualbox-go"
)

func main() {
	name := "testVM"
	vb, vm := CreateVM(name)
	vm, err := GetVMInfo(name)
	if err != nil {
		fmt.Errorf("Get info VM failed: %s", err.Error())
	}
	vb.DeleteVM(vm)

}
func CreateVM(name string) (*vbg.VBox, *vbg.VirtualMachine) {
	// setup temp directory, that will be used to cache different VM related files during the creation of the VM.
	dirName, err := ioutil.TempDir("", "vbm")
	if err != nil {
		fmt.Errorf("Tempdir creation failed %v", err)
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
		fmt.Errorf("CreateDisk failed %v", err)
	}

	vm := &vbg.VirtualMachine{}
	vm.Spec.Name = name
	vm.Spec.OSType = vbg.Linux64
	vm.Spec.CPU.Count = 2
	vm.Spec.Memory.SizeMB = 1000
	vm.Spec.Disks = []vbg.Disk{disk1}

	err = vb.CreateVM(vm)
	if err != nil {
		fmt.Errorf("Failed creating vm %v", err)
	}

	err = vb.RegisterVM(vm)
	if err != nil {
		fmt.Errorf("Failed registering vm")
	}

	return vb, vm
}

func GetVMInfo(name string) (machine *vbg.VirtualMachine, err error) {
	vb := vbg.NewVBox(vbg.Config{})
	return vb.VMInfo(name)
}
