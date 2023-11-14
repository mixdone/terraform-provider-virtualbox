package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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

	info, err := vb.VMInfo(vm.Spec.Name)
	if err != nil {
		logrus.Fatalf("Failed VMInfo %v", err.Error())
	}

	if info.Spec.Name != vm.Spec.Name ||
		info.Spec.OSType != vm.Spec.OSType ||
		info.Spec.CPU != vm.Spec.CPU ||
		info.Spec.Memory != vm.Spec.Memory ||
		info.Spec.Disks[0] != vm.Spec.Disks[0] {
		logrus.Fatalf(
			"Expected some fields to be auto created, have %v %v %v %v %+v",
			info.Spec.Name,
			info.Spec.OSType,
			info.Spec.CPU,
			info.Spec.Memory,
			info.Spec.Disks,
		)
	}
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

	ctx := context.Background()             // создаём фоновый контекст
	context.WithTimeout(ctx, 1*time.Minute) // с таймаутом в 1 минуту

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

	ctx := context.Background()             // создаём фоновый контекст
	context.WithTimeout(ctx, 1*time.Minute) // с таймаутом в 1 минуту

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
