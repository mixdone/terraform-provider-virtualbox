package pkg

import (
	"fmt"
	"path/filepath"

	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"
)

type LoadingType int

// now here 3 types of loading
const (
	vdiLoading LoadingType = iota
	imageloading
	empty
)

type VMConfig struct {
	Name       string
	CPUs       int
	Memory     int
	Image_path string
	Dirname    string
	Ltype      LoadingType
	Vdi_size   int64
	OS_id      string
	Group      string
	Snapshot   vbg.Snapshot
	NICs       []vbg.NIC
}

// create VM with chosen loading type

func CreateVM(vmCfg VMConfig) (*vbg.VirtualMachine, error) {
	// make path to existing vdi or create name from new vdi
	var vdiDisk string
	switch vmCfg.Ltype {
	case vdiLoading:
		vdiDisk = vmCfg.Image_path
	case imageloading:
		vdiDisk = filepath.Base(vmCfg.Image_path)
		vdiDisk = vdiDisk[:len(vdiDisk)-len(filepath.Ext(vdiDisk))] + vmCfg.Name + ".vdi"
		vdiDisk = filepath.Join(vmCfg.Dirname, vdiDisk)
	}

	vb := vbg.NewVBox(vbg.Config{
		BasePath: vmCfg.Dirname,
	})

	storageController1 := vbg.StorageController{
		Name: "SATA Controller",
		Type: vbg.SATA,
	}

	sata := vbg.StorageControllerAttachment{
		Type:   vbg.SATA,
		Name:   "SATA Controller",
		Port:   0,
		Device: 0,
	}

	disk_VDI := vbg.Disk{
		Path:       vdiDisk,
		Format:     vbg.VDI,
		SizeMB:     vmCfg.Vdi_size,
		Type:       "hdd",
		Controller: sata,
	}

	// Loader for the image
	storageController2 := vbg.StorageController{
		Name: "IDE Controller",
		Type: vbg.IDE,
	}

	ide := vbg.StorageControllerAttachment{
		Type:   vbg.IDE,
		Name:   "IDE Controller",
		Port:   1,
		Device: 0,
	}

	disk_ISO := vbg.Disk{
		Path:       vmCfg.Image_path,
		Type:       "dvddrive",
		Controller: ide,
	}

	if vmCfg.Ltype == imageloading {
		if err := vb.CreateDisk(&disk_VDI); err != nil {
			return nil, fmt.Errorf("disk creation failed: %s", err.Error())
		}
	}

	var disks []vbg.Disk
	switch vmCfg.Ltype {
	case vdiLoading:
		disks = []vbg.Disk{disk_VDI}
	case imageloading:
		disks = []vbg.Disk{disk_VDI, disk_ISO}
	case empty:
		disks = []vbg.Disk{}
	}

	// Parameters of the virtual machine
	spec := &vbg.VirtualMachineSpec{
		Name:            vmCfg.Name,
		OSType:          vbg.OSType{ID: vmCfg.OS_id},
		CPU:             vbg.CPU{Count: vmCfg.CPUs},
		Memory:          vbg.Memory{SizeMB: vmCfg.Memory},
		Disks:           disks,
		Group:           vmCfg.Group,
		CurrentSnapshot: vmCfg.Snapshot,
		NICs:            vmCfg.NICs,
	}

	vm := &vbg.VirtualMachine{
		Spec: *spec,
	}

	logrus.Infoln("Creating VM with CPU and memory", vm.Spec.CPU, vm.Spec.Memory)

	if err := vb.CreateVM(vm); err != nil {
		return nil, fmt.Errorf("VM creation failed: %s", err.Error())
	}

	if err := vb.RegisterVM(vm); err != nil {
		return nil, fmt.Errorf("failed registering vm: %s", err.Error())
	}

	// Set CPUs and memory
	if err := vb.SetCPUCount(vm, vm.Spec.CPU.Count); err != nil {
		return nil, fmt.Errorf("set CPU Count failed: %s", err.Error())
	}

	if err := vb.SetMemory(vm, vm.Spec.Memory.SizeMB); err != nil {
		return nil, fmt.Errorf("set memory failed: %s", err.Error())
	}

	if err := vb.ModifyVM(vm, []string{"network_adapter"}); err != nil {
		return nil, fmt.Errorf("set network failed: %s", err.Error())
	}

	// Connecting a disk to a virtual machine
	if vmCfg.Ltype != empty {
		if err := vb.AddStorageController(vm, storageController1); err != nil {
			return nil, fmt.Errorf("add SATA controller error: %s", err.Error())
		}

		if err := vb.AttachStorage(vm, &disk_VDI); err != nil {
			return nil, fmt.Errorf("attach error: %s", err.Error())
		}
	}

	if vmCfg.Ltype == imageloading {
		// Connecting the installation disk image
		if err := vb.AddStorageController(vm, storageController2); err != nil {
			return nil, fmt.Errorf("add IDE controller error: %s", err.Error())
		}

		if err := vb.AttachStorage(vm, &disk_ISO); err != nil {
			return nil, fmt.Errorf("attach error: %s", err.Error())
		}
	}

	if vm.Spec.CurrentSnapshot.Name != "" {
		err := vb.TakeSnapshot(vm, vm.Spec.CurrentSnapshot, false)
		if err != nil {
			return nil, fmt.Errorf("TakeSnapshot failed: %s", err.Error())
		}
	}

	return vm, nil
}
