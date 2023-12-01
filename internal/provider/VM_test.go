package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mixdone/terraform-provider-virtualbox/internal/provider/pkg"
	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

func Test_createVM(t *testing.T) {

	logrus.Info("setup")
	dir, _ := os.UserHomeDir()

	dirName, err := os.MkdirTemp(dir, "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{
		BasePath: dirName,
	})

	disk1 := vbg.Disk{
		Path:   filepath.Join(dirName, "disk1.vdi"),
		Format: vbg.VDI,
		SizeMB: 1000,
	}

	err = vb.CreateDisk(&disk1)
	if err != nil {
		logrus.Fatalf("CreateDisk failed %v", err.Error())
	}

	// Параметры вируальной машины
	spec := &vbg.VirtualMachineSpec{
		Name:   "vmName",
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: 2},
		Memory: vbg.Memory{SizeMB: 1000},
		Disks:  []vbg.Disk{disk1},
	}
	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	if err := vb.CreateVM(vm); err != nil {
		logrus.Fatalf("Failed creating %v", err.Error())
	}

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Fatalf("Failed register %v", err.Error())
	}

	/*info, err := vb.VMInfo(vm.Spec.Name)
	if err != nil {
		logrus.Fatalf("Failed VMInfo %v", err.Error())
	}

	if info.Spec.Name != vm.Spec.Name {
		logrus.Fatalf("Expected name: %v, actual name: %v", vm.Spec.Name, info.Spec.Name)
	}
	if info.Spec.OSType.ID != vm.Spec.OSType.ID {
		logrus.Fatalf("Expected OS: %v, actual OS: %v", vm.Spec.OSType.ID, info.Spec.OSType.ID)
	}
	if info.Spec.CPU.Count != vm.Spec.CPU.Count {
		logrus.Fatalf("Expected cpu count: %v, actual cpu count: %v", vm.Spec.CPU.Count, info.Spec.CPU.Count)
	}
	if info.Spec.Memory.SizeMB != vm.Spec.Memory.SizeMB {
		logrus.Fatalf("Expected memory: %v, actual memory: %v", vm.Spec.Memory, info.Spec.Memory)
	}
	if info.Spec.Disks[0].Path != vm.Spec.Disks[0].Path {
		logrus.Fatalf("Expected disk path: %v, actual disk path: %v", vm.Spec.Disks[0].Path, info.Spec.Disks[0].Path)
	}
	if info.Spec.Disks[0].SizeMB != vm.Spec.Disks[0].SizeMB {
		logrus.Fatalf("Expected disk size: %v, actual disk size: %v", vm.Spec.Disks[0].SizeMB, info.Spec.Disks[0].SizeMB)
	}
	if string(info.Spec.Disks[0].Format) != string(vm.Spec.Disks[0].Format) {
		logrus.Fatalf("Expected disk format: %v, actual disk format: %v", string(vm.Spec.Disks[0].Format), string(info.Spec.Disks[0].Format))
	}*/
}

func Test_define(t *testing.T) {

	dir, _ := os.UserHomeDir()
	dirName, err := os.MkdirTemp(dir, "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{})
	disk1 := vbg.Disk{
		Path:   filepath.Join(dirName, "disk1.vdi"),
		Format: vbg.VDI,
		SizeMB: 10000,
		Type:   "hdd",
		Controller: vbg.StorageControllerAttachment{
			Type: vbg.IDE,
		},
	}

	// Параметры вируальной машины
	spec := &vbg.VirtualMachineSpec{
		Name:   "vmName",
		Group:  "/vmGroup",
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: 2},
		Memory: vbg.Memory{SizeMB: 1000},
		Disks:  []vbg.Disk{disk1},
	}
	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	_, err = vb.EnsureDefaults(vm) // проверка на то, что ожидаемые значения доступны для правильного выполнения функции
	if err != nil {
		logrus.Fatalf("Problem with defaults %v", err.Error())
	}

	ctx := context.Background() // создаём фоновый контекст

	vb.UnRegisterVM(vm)
	vb.DeleteVM(vm)

	defer func(vb *vbg.VBox, vm *vbg.VirtualMachine) {
		err := vb.DeleteVM(vm)
		if err != nil {
			logrus.Fatalf("Problem with deleting VM %v", err.Error())
		}
	}(vb, vm)
	defer func(vb *vbg.VBox, vm *vbg.VirtualMachine) {
		err := vb.UnRegisterVM(vm)
		if err != nil {
			logrus.Fatalf("Problem with cancellation of registration of VM %v", err.Error())
		}
	}(vb, vm)

	nvm, err := vb.Define(ctx, vm)
	if err != nil {
		logrus.Fatalf("Error %+v", err.Error())
	}

	fmt.Printf("Created %#v\nCreated %#v\n", vm, nvm)
}

func Test_states(t *testing.T) {
	dir, _ := os.UserHomeDir()
	dirName, err := os.MkdirTemp(dir, "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	vb := vbg.NewVBox(vbg.Config{})
	disk1 := vbg.Disk{
		Path:   filepath.Join(dirName, "disk1.vdi"),
		Format: vbg.VDI,
		SizeMB: 1000,
		Type:   "hdd",
		Controller: vbg.StorageControllerAttachment{
			Type: vbg.IDE,
		},
	}

	// Параметры вируальной машины
	spec := &vbg.VirtualMachineSpec{
		Name:   "vmName",
		Group:  "/vmGroup",
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: 2},
		Memory: vbg.Memory{SizeMB: 1000},
		Disks:  []vbg.Disk{disk1},
		State:  vbg.Running,
	}
	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	_, err = vb.EnsureDefaults(vm) // проверка на то, что ожидаемые значения доступны для правильного выполнения функции
	if err != nil {
		logrus.Fatalf("Problem with defaults %v", err.Error())
	}

	ctx := context.Background() // создаём фоновый контекст

	vb.UnRegisterVM(vm)
	vb.DeleteVM(vm)

	nvm, err := vb.Define(ctx, vm)

	if err != nil {
		logrus.Fatalf("%v", err.Error())
	} else if nvm.UUID == "" {
		logrus.Fatalf("VM is not discoverable after creation %s", vm.Spec.Name)
	}

	if _, err = vb.Start(vm); err != nil {
		logrus.Fatalf("Failed to start %s: error %v", vm.Spec.Name, err.Error())
	}

	if _, err = vb.Stop(vm); err != nil {
		logrus.Fatalf("Failed to stop %s: error %v", vm.Spec.Name, err.Error())
	}
}

func Test_CreatePath(t *testing.T) {

	logrus.Info("setup")

	vb := vbg.NewVBox(vbg.Config{}) // путь не указан

	// Параметры вируальной машины
	spec := &vbg.VirtualMachineSpec{
		Name:   "vmName",
		OSType: vbg.Ubuntu64,
		CPU:    vbg.CPU{Count: 2},
		Memory: vbg.Memory{SizeMB: 1000},
	}
	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	if err := vb.CreateVM(vm); err != nil {
		logrus.Fatalf("Failed creating %v", err.Error())
	}

	if err := vb.RegisterVM(vm); err != nil {
		logrus.Fatalf("Failed register %v", err.Error())
	}

	if err := vb.UnRegisterVM(vm); err != nil {
		logrus.Fatalf("Failed unregister %v", err.Error())
	}

	if err := vb.DeleteVM(vm); err != nil {
		logrus.Fatalf("Failed delete %v", err.Error())
	}
}

func Test_ControlVM(t *testing.T) {
	name := "test_ControlVM"
	memory := 1024
	cpus := 2
	url := "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
	basedir := "VMS1"
	homedir, _ := os.UserHomeDir()
	machinesDir := filepath.Join(homedir, basedir)
	installedData := filepath.Join(homedir, "InstalledData")
	var ltype pkg.LoadingType = 2

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		logrus.Fatalf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		logrus.Fatalf("Creation InstalledData foldier failed: %s", err.Error())
	}

	vm, err := pkg.CreateVM(name, cpus, memory, url, machinesDir, ltype)
	if err != nil {
		logrus.Fatalf("Creation VM failed: %s", err.Error())
	}

	vb := vbg.NewVBox(vbg.Config{BasePath: machinesDir})

	if _, err = vb.ControlVM(vm, "running"); err != nil {
		logrus.Fatalf("Failed running %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "pause"); err != nil {
		logrus.Fatalf("Failed pause %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "resume"); err != nil {
		logrus.Fatalf("Failed resume %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "reset"); err != nil {
		logrus.Fatalf("Failed reset %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "save"); err != nil {
		logrus.Fatalf("Failed save %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "running"); err != nil {
		logrus.Fatalf("Failed running %v", err.Error())
	}

	if _, err = vb.ControlVM(vm, "poweroff"); err != nil {
		logrus.Fatalf("Failed poweroff %v", err.Error())
	}

	defer vb.DeleteVM(vm)
	defer vb.UnRegisterVM(vm)
}

func Test_ModifyVM(t *testing.T) {
	name := "test_ModifyVM"
	memory := 1024
	cpus := 2
	url := "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
	basedir := "VMS1"
	homedir, _ := os.UserHomeDir()
	machinesDir := filepath.Join(homedir, basedir)
	installedData := filepath.Join(homedir, "InstalledData")
	var ltype pkg.LoadingType = 2

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		logrus.Fatalf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		logrus.Fatalf("Creation InstalledData foldier failed: %s", err.Error())
	}

	vm, err := pkg.CreateVM(name, cpus, memory, url, machinesDir, ltype)
	if err != nil {
		logrus.Fatalf("Creation VM failed: %s", err.Error())
	}

	vb := vbg.NewVBox(vbg.Config{BasePath: machinesDir})

	vm.Spec.Memory.SizeMB = 512
	vm.Spec.CPU.Count = 1
	vm.Spec.OSType.ID = "Ubuntu_64"
	if err = vb.ModifyVM(vm, []string{"memory", "cpus", "ostype"}); err != nil {
		logrus.Fatalf("ModifyVM failed: %s", err.Error())
	}

	vb = vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, basedir)})
	vm2, _ := vb.VMInfo(vm.Spec.Name)
	if vm.Spec.CPU.Count != vm2.Spec.CPU.Count {
		logrus.Fatalf("CPU count has not been changed")
	}
	if vm.Spec.Memory.SizeMB != vm2.Spec.Memory.SizeMB {
		logrus.Fatalf("Memory has not been changed")
	}

	defer vb.DeleteVM(vm)
	defer vb.UnRegisterVM(vm)
}
