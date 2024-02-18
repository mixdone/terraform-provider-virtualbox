package main

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"

	vbg "github.com/mixdone/virtualbox-go"
)

func main() {

	vmName, CPUs, memory := "terraformVM_testVMnetwork_adapter", 1, 1024
	CreateVM(vmName, CPUs, memory)

}

func CreateVM(vmName string, CPUs, memory int) {
	dirName, err := os.MkdirTemp("./", "VirtualBox VMs")
	if err != nil {
		logrus.Fatalf("Tempdir creation failed %v", err.Error())
	}
	defer os.RemoveAll(dirName)

	config := vbg.Config{
		BasePath: dirName,
	}

	vb := vbg.NewVBox(config)

	//vbg.NWMode_natnetwork, vbg.NWMode_intnet - не работают

	array := []vbg.NetworkMode{
		vbg.NWMode_none,
		vbg.NWMode_null,
		vbg.NWMode_nat,
		vbg.NWMode_bridged,
		vbg.NWMode_hostonly,
		vbg.NWMode_generic,
	}

	arr := []vbg.NICType{vbg.NIC_Am79C970A,
		vbg.NIC_Am79C973,
		vbg.NIC_82540EM,
		vbg.NIC_82543GC,
		vbg.NIC_82545EM,
		vbg.NIC_virtio}

	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			NIC1 := &vbg.NIC{
				Index:          3,
				Mode:           array[i],
				Type:           arr[j],
				CableConnected: true,
			}

			spec := &vbg.VirtualMachineSpec{
				Name:   vmName + strconv.Itoa(i) + strconv.Itoa(j),
				OSType: vbg.Ubuntu64,
				CPU:    vbg.CPU{Count: CPUs},
				Memory: vbg.Memory{SizeMB: memory},
				NICs:   []vbg.NIC{*NIC1},
			}

			vm := &vbg.VirtualMachine{
				Spec: *spec,
			}

			err = vb.CreateVM(vm)
			if err != nil {
				logrus.Fatalf("VM creation failed: %s", err.Error())
			}

			err = vb.RegisterVM(vm)
			if err != nil {
				logrus.Fatalf("Failed registering vm")
			}

			vb.SetCPUCount(vm, vm.Spec.CPU.Count)
			vb.SetMemory(vm, vm.Spec.Memory.SizeMB)
			vb.ModifyVM(vm, []string{"network_adapter"})

		}
	}

}

func GetVMInfo(name string, config vbg.Config) (*vbg.VirtualMachine, error) {
	vb := vbg.NewVBox(config)
	vm, err := vb.VMInfo(name)
	return vm, err
}
