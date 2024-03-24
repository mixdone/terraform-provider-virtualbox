package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mixdone/terraform-provider-virtualbox/pkg"
	vbg "github.com/mixdone/virtualbox-go"
	"github.com/sirupsen/logrus"

	mem "github.com/pbnjay/memory"
)

func resourceVM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVirtualBoxCreate,
		ReadContext:   resourceVirtualBoxRead,
		UpdateContext: resourceVirtualBoxUpdate,
		DeleteContext: resourceVirtualBoxDelete,
		Exists:        resourceVirtualBoxExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Virtual Machine name.",
				Type:        schema.TypeString,
				Required:    true,
			},

			"basedir": {
				Description: "The folder in which the virtual machine data will be located.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "VMs",
				ForceNew:    true,
			},

			"memory": {
				Description: "RAW allocated for machine.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     128,
			},

			"disk_size": {
				Description: "VDI size in MB.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     15000,
			},

			"group": {
				Description: "Group of Virtual Machines.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},

			"cpus": {
				Description: "Amount of CPUs.",
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     2,
			},

			"status": {
				Description: "Status of Virtual Machine.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "poweroff",
			},

			"image": {
				Description: "Path to image that is located on the host.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"url": {
				Description: "The link from which the image or disk will be downloaded.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},

			"disk": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_adapter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"index": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"network_mode": {
							Description: "nat, hostonly etc",
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "none",
						},
						"nic_type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "Am79C970A",
						},
						"cable_connected": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},

			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"os_id": {
				Description: "Specifies the guest OS to run in the VM.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Linux_64",
			},

			"snapshot": {
				Type:        schema.TypeList,
				Description: "Adds a list of snapshots. You can add a new Snapshot, edit or delete existing ones.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},

						"description": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
					},
				},
			},
		},
	}
}

func resourceVirtualBoxCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Geting data from config
	if err := validateVmParams(d); err != nil {
		return diag.Errorf(err.Error())
	}

	var vmConf pkg.VMConfig

	vmConf.Name = d.Get("name").(string)
	vmConf.CPUs = d.Get("cpus").(int)
	vmConf.Memory = d.Get("memory").(int)
	vmConf.DiskSize = int64(d.Get("disk_size").(int))
	vmConf.OS_id = d.Get("os_id").(string)
	vmConf.Group = d.Get("group").(string)

	snapshots := d.Get("snapshot.#").(int)
	if snapshots > 0 {
		snapshots = 1
		req1 := fmt.Sprintf("snapshot.%d.name", snapshots-1)
		req2 := fmt.Sprintf("snapshot.%d.description", snapshots-1)
		vmConf.Snapshot.Name = d.Get(req1).(string)
		vmConf.Snapshot.Description = d.Get(req2).(string)
	} else {
		vmConf.Snapshot = vbg.Snapshot{Name: "", Description: ""}
	}

	// Making new folders for VirtualMachine data
	homedir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	machinesDir := filepath.Join(homedir, d.Get("basedir").(string))
	installedData := filepath.Join(machinesDir, "InstalledData")

	vmConf.Dirname = machinesDir

	if err := os.MkdirAll(machinesDir, 0740); err != nil {
		return diag.Errorf("Creation VirtualMachines foldier failed: %s", err.Error())
	}
	if err := os.MkdirAll(installedData, 0740); err != nil {
		return diag.Errorf("Creation InstalledData foldier failed: %s", err.Error())
	}

	var ltype pkg.LoadingType

	// Obtaining the image
	im, ok := d.GetOk("image")
	image := im.(string)
	if !ok {
		url, ok := d.GetOk("url")
		if !ok {
			disk, ok := d.GetOk("disk")
			if !ok {
				ltype = 2
			}
			image = disk.(string)
		} else {
			filename, err := pkg.FileDownload(url.(string), homedir)
			if err != nil {
				return diag.Errorf("File dowload failed: %s", err.Error())
			}

			if filepath.Ext(filepath.Base(filename)) != ".iso" {
				imagePath, err := pkg.UnpackImage(filename, installedData)
				if err != nil {
					return diag.Errorf("File unpaking failed: %s", err.Error())
				}
				basePath := filepath.Base(imagePath)
				basePath = basePath[:len(basePath)-len(filepath.Ext(basePath))] + vmConf.Name + filepath.Ext(basePath)
				image = filepath.Join(imagePath[:len(imagePath)-len(filepath.Base(imagePath))], basePath)
				os.Rename(imagePath, image)
			} else {
				image = filename
			}
		}
	} else {
		if filepath.Ext(filepath.Base(image)) == ".iso" {
			ltype = 1
		} else {
			ltype = 0
		}
	}

	var NICs [20]vbg.NIC

	for i, nic := range NICs {
		nic.Index = i
		nic.NetworkName = ""
		nic.Mode = "none"
		nic.Type = "Am79C970A"
		nic.CableConnected = false
	}
	//rule := make([]vbg.PortForwarding, 10)
	nicNumber := d.Get("network_adapter.#").(int)

	for i := 0; i < nicNumber; i++ {

		requestMode := fmt.Sprintf("network_adapter.%d.network_mode", i)
		currentMode := d.Get(requestMode).(string)

		requestType := fmt.Sprintf("network_adapter.%d.nic_type", i)
		currentType := d.Get(requestType).(string)

		requestCable := fmt.Sprintf("network_adapter.%d.cable_connected", i)
		currentCable := d.Get(requestCable).(bool)

		NICs[i].Index = i + 1
		NICs[i].Mode = vbg.NetworkMode(currentMode)
		NICs[i].Type = vbg.NICType(currentType)
		NICs[i].CableConnected = currentCable
	}

	vmConf.Ltype = ltype
	vmConf.Image_path = image
	vmConf.NICs = NICs[:]

	// Creating VM with specified parametrs
	vm, err := pkg.CreateVM(vmConf)
	if err != nil {
		return diag.Errorf("Creation VM failed: %s", err.Error())
	}

	// Setting the VM id for Terraform
	d.SetId(vm.UUIDOrName())

	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})

	vm, err = vb.VMInfo(d.Id())
	if err != nil {
		d.SetId("")
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	status := d.Get("status").(string)

	if status != "poweroff" {
		if _, err := vb.ControlVM(vm, status); err != nil {
			return diag.Errorf("Unable to running VM: %s", err.Error())
		}
		vm.Spec.State = vbg.VirtualMachineState(status)
		if err = setState(d, vm); err != nil {
			return diag.Errorf("Setting state failed: %s", err.Error())
		}
	}

	return resourceVirtualBoxRead(ctx, d, m)
}

func resourceVirtualBoxRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Getting Machine by id
	homedir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	basePath := filepath.Join(homedir, d.Get("basedir").(string))
	vb := vbg.NewVBox(vbg.Config{BasePath: basePath})
	vm, err := vb.VMInfo(d.Id())

	logrus.Info(vm.Spec.CurrentSnapshot.Name)

	for i := 0; i < len(vm.Spec.Snapshots); i++ {
		logrus.Info(vm.Spec.Snapshots[i].Name)
		logrus.Info(vm.Spec.Snapshots[i].Description)
	}

	if err != nil {
		d.SetId("")
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	// Set state of Machine for Terraform
	if err := setState(d, vm); err != nil {
		return diag.Errorf("Didn't manage to set VMState: %s", err.Error())
	}

	if err := setNetwork(d, vm); err != nil {
		return diag.Errorf("Didn't manage to set VMState: %s", err.Error())
	}

	// Set name of Machine for Terraform
	if err := d.Set("name", vm.Spec.Name); err != nil {
		return diag.Errorf("Didn't manage to set name: %s", err.Error())
	}

	arr := make([]map[string]interface{}, 0, 3)

	for i := 0; i < len(vm.Spec.Snapshots); i++ {
		arr = append(arr, map[string]interface{}{
			"name":        vm.Spec.Snapshots[i].Name,
			"description": vm.Spec.Snapshots[i].Description,
		})
	}

	if err := d.Set("snapshot", arr); err != nil {
		return diag.Errorf("Didn't manage to set snapshot: %s", err.Error())
	}

	// Set CPUs amount for Terraform
	if err := d.Set("cpus", vm.Spec.CPU.Count); err != nil {
		return diag.Errorf("Didn't manage to set cpus: %s", err.Error())
	}

	// Set memory for Terraform
	if err := d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		return diag.Errorf("Didn't manage to set memory: %s", err.Error())
	}

	// Set basedir VM for Terraform
	if err := d.Set("basedir", d.Get("basedir").(string)); err != nil {
		return diag.Errorf("Didn't manage to set basedir: %s", err.Error())
	}

	return nil
}

func poweroffVM(ctx context.Context, d *schema.ResourceData, vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	if _, err := vb.ControlVM(vm, "poweroff"); err != nil {
		logrus.Errorf("Unable to poweroff VM: %s", err.Error())
		return err
	}

	vm.Spec.State = vbg.Poweroff
	return nil
}

func resourceVirtualBoxUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if err := validateVmParams(d); err != nil {
		return diag.Errorf(err.Error())
	}

	// Getting VM by id
	homedir, _ := os.UserHomeDir()
	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})
	vm, err := vb.VMInfo(d.Id())

	// Array of parametrs
	parameters := []string{}

	if err != nil {
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	// Powerof VM
	if err = poweroffVM(ctx, d, vm, vb); err != nil {
		return diag.Errorf("Setting state failed: %s", err.Error())
	}

	// Setting new name
	actualName := vm.Spec.Name
	newName := d.Get("name").(string)
	if actualName != newName {
		parameters = append(parameters, "name")
		vm.Spec.Name = newName
	}

	// Setting new amount of memory
	actualMemory := vm.Spec.Memory.SizeMB
	newMemory := d.Get("memory").(int)
	if actualMemory != newMemory {
		parameters = append(parameters, "memory")
		vm.Spec.Memory.SizeMB = newMemory
	}

	// Setting new amount of CPUs
	actualCPUCount := vm.Spec.CPU.Count
	newCPUCount := d.Get("cpus").(int)
	if actualCPUCount != newCPUCount {
		parameters = append(parameters, "cpus")
		vm.Spec.CPU.Count = newCPUCount
	}

	// Setting new os type
	actualOs_id := vm.Spec.OSType.ID
	newOs_id := d.Get("os_id").(string)
	if actualOs_id != newOs_id {
		parameters = append(parameters, "ostype")
		vm.Spec.OSType.ID = newOs_id
	}

	// Setting new network adapters
	needAppendNetwork := false
	nicNumber := d.Get("network_adapter.#").(int)

	if len(vm.Spec.NICs) < nicNumber {
		var NICs = make([]vbg.NIC, nicNumber-len(vm.Spec.NICs))
		vm.Spec.NICs = append(vm.Spec.NICs, NICs...)
	}

	for i := 0; i < nicNumber; i++ {
		vm.Spec.NICs[i].Index = i + 1
		requestMode := fmt.Sprintf("network_adapter.%d.network_mode", i)
		currentMode := vbg.NetworkMode(d.Get(requestMode).(string))
		if currentMode != vm.Spec.NICs[i].Mode {
			needAppendNetwork = true
			vm.Spec.NICs[i].Mode = currentMode
		}

		requestType := fmt.Sprintf("network_adapter.%d.nic_type", i)
		currentType := vbg.NICType(d.Get(requestType).(string))
		if currentType != vm.Spec.NICs[i].Type {
			needAppendNetwork = true
			vm.Spec.NICs[i].Type = currentType
		}

		requestCable := fmt.Sprintf("network_adapter.%d.cable_connected", i)
		currentCable := d.Get(requestCable).(bool)
		if currentCable != vm.Spec.NICs[i].CableConnected {
			needAppendNetwork = true
			vm.Spec.NICs[i].CableConnected = currentCable
		}
	}
	if needAppendNetwork {
		parameters = append(parameters, "network_adapter")
	}

	group := d.Get("group").(string)
	if vm.Spec.Group != group {
		parameters = append(parameters, "group")
		vm.Spec.Group = group
	}

	// Modify VM
	if len(parameters) != 0 {
		err = vb.ModifyVM(vm, parameters)
		if err != nil {
			return diag.Errorf("ModifyVM failed: %s", err.Error())
		}
	}

	// Setting new VM id
	vm.UUID = vm.UUIDOrName()
	d.SetId(vm.UUIDOrName())

	// Updating state
	status := d.Get("status").(string)

	logrus.Printf("%s -> %s", vm.Spec.State, status)
	if status != string(vm.Spec.State) {
		if _, err := vb.ControlVM(vm, status); err != nil {
			return diag.Errorf("Unable to running VM: %s", err.Error())
		}
		logrus.Printf("%s -> %s", vm.Spec.State, status)
		vm.Spec.State = vbg.VirtualMachineState(status)
		if err = setState(d, vm); err != nil {
			return diag.Errorf("Setting state failed: %s", err.Error())
		}
	}

	snapshots := d.Get("snapshot.#").(int)

	diff := len(vm.Spec.Snapshots) - snapshots

	if diff < 0 {
		emptySnapshts := make([]vbg.Snapshot, -diff)
		vm.Spec.Snapshots = append(vm.Spec.Snapshots, emptySnapshts...)
	}

	for i := 0; i < snapshots; i++ {
		var snapshot vbg.Snapshot
		req1 := fmt.Sprintf("snapshot.%d.name", i)
		req2 := fmt.Sprintf("snapshot.%d.description", i)

		snapshot.Name = d.Get(req1).(string)
		snapshot.Description = d.Get(req2).(string)

		if vm.Spec.Snapshots[i].Name == "" {
			snapshotOperationsHandler(vb, vm, vm.Spec.Snapshots[i], snapshot, "take", status)
		} else if snapshot.Name != vm.Spec.Snapshots[i].Name ||
			snapshot.Description != vm.Spec.Snapshots[i].Description {
			snapshotOperationsHandler(vb, vm, vm.Spec.Snapshots[i], snapshot, "update", status)
		}
	}

	if diff > 0 {
		for i := snapshots; i < snapshots+diff; i++ {
			snapshotOperationsHandler(vb, vm, vm.Spec.Snapshots[i], vm.Spec.Snapshots[i], "delete", status)
		}
	}

	return resourceVirtualBoxRead(ctx, d, m)
}

func snapshotOperationsHandler(vb *vbg.VBox, vm *vbg.VirtualMachine, prevSnapshot vbg.Snapshot, snapshot vbg.Snapshot, operation string, status string) error {
	var err error
	switch operation {
	case "take":
		if snapshot.Name != vm.Spec.CurrentSnapshot.Name {
			if status == "running" {
				err = vb.TakeSnapshot(vm, snapshot, true)
			} else {
				err = vb.TakeSnapshot(vm, snapshot, false)
			}
		}
		return err
	case "delete":
		return vb.DeleteSnapshot(vm, snapshot)
	case "update":
		if snapshot.Name != "" {
			err = vb.EditSnapshot(vm, prevSnapshot, snapshot)
		}
		return err
	case "restore":
		if snapshot.Name != "" {
			err = vb.RestoreSnapshot(vm, snapshot)
		}
		return err
	default:
		return fmt.Errorf("unknown snapshot operation\nUsage: operation=[take|delete|update|restore]")
	}
}

func resourceVirtualBoxExists(d *schema.ResourceData, m interface{}) (bool, error) {
	vb := vbg.NewVBox(vbg.Config{})
	_, err := vb.VMInfo(d.Id())
	switch err {
	case nil:
		return true, nil
	case vbg.ErrMachineNotExist:
		return false, nil
	default:
		return false, fmt.Errorf("VMInfo failed: %s", err)
	}
}

func resourceVirtualBoxDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Getting VM by id
	homedir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})
	vm, err := vb.VMInfo(d.Id())
	if err != nil {
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	// Powerof VM
	if err = poweroffVM(ctx, d, vm, vb); err != nil {
		return diag.Errorf("Setting state failed: %s", err.Error())
	}

	// Unresitering VM
	if err = vb.UnRegisterVM(vm); err != nil {
		return diag.Errorf("VM Unregister failed: %s", err.Error())
	}

	// VM deletion
	if err = vb.DeleteVM(vm); err != nil {
		return diag.Errorf("VM deletion failed: %s", err.Error())
	}

	// Delete machine folder
	machineDir := filepath.Join(homedir, d.Get("basedir").(string))
	if err := os.RemoveAll(machineDir); err != nil {
		return diag.Errorf("Can't clear the data: %s", err.Error())
	}

	return nil
}

func setState(d *schema.ResourceData, vm *vbg.VirtualMachine) error {
	var err error
	switch vm.Spec.State {
	case vbg.Poweroff:
		err = d.Set("status", "poweroff")
	case vbg.Running:
		err = d.Set("status", "running")
	case vbg.Paused:
		err = d.Set("status", "paused")
	case vbg.Saved:
		err = d.Set("status", "saved")
	case vbg.Aborted:
		err = d.Set("status", "aborted")
	}
	return err
}

func setNetwork(d *schema.ResourceData, vm *vbg.VirtualMachine) error {
	getType := func(nic vbg.NIC) string {
		switch nic.Type {
		case vbg.NIC_Am79C970A:
			return "Am79C970A"
		case vbg.NIC_Am79C973:
			return "Am79C973"
		case vbg.NIC_82540EM:
			return "82540EM"
		case vbg.NIC_82543GC:
			return "82543GC"
		case vbg.NIC_82545EM:
			return "82545EM"
		case vbg.NIC_virtio:
			return "virtio"
		default:
			return ""
		}
	}

	getMode := func(nic vbg.NIC) string {
		switch nic.Mode {
		case vbg.NWMode_none:
			return "none"
		case vbg.NWMode_null:
			return "null"
		case vbg.NWMode_nat:
			return "nat"
		case vbg.NWMode_natnetwork:
			return "natnetwork"
		case vbg.NWMode_bridged:
			return "bridged"
		case vbg.NWMode_intnet:
			return "intnet"
		case vbg.NWMode_hostonly:
			return "hostonly"
		case vbg.NWMode_generic:
			return "generic"
		default:
			return ""
		}
	}

	nics := make([]map[string]any, 0, 4)
	for i, nic := range vm.Spec.NICs {
		out := make(map[string]any)
		out["index"] = i + 1
		out["network_mode"] = getMode(nic)
		out["nic_type"] = getType(nic)
		out["cable_connected"] = nic.CableConnected
		nics = append(nics, out)
	}

	if err := d.Set("network_adapter", nics); err != nil {
		return err
	}

	return nil
}

func validateVmParams(d *schema.ResourceData) error {
	amountOfProblems := 0
	var error_output []string

	cpus := d.Get("cpus").(int)
	if cpus <= 0 || cpus >= runtime.NumCPU() {
		error_output = append(error_output, fmt.Sprintf("Set the number of CPUs according to the following limits: 1 - %v", runtime.NumCPU()))
		amountOfProblems++
	}

	memory := d.Get("memory").(int)
	if memory <= 0 || memory > int(mem.TotalMemory()) {
		error_output = append(error_output, fmt.Sprintf("Set the amount of memory according to the following limits: 1 - %v", mem.TotalMemory()))
		amountOfProblems++
	}

	status := d.Get("status").(string)
	switch status {
	case "poweroff":
		break
	case "running":
		break
	case "paused":
		break
	case "saved":
		break
	case "aborted":
		break
	default:
		error_output = append(error_output, "Status does not match any of the existing ones\n - poweroff\n - runnning\n - paused \n - saved \n - aborted")
		amountOfProblems++
	}

	if amountOfProblems == 0 {
		return nil
	}

	report := "\n"
	for i := 1; i <= amountOfProblems; i++ {
		report += fmt.Sprintf("%v) %s\n", i, error_output[i-1])
	}
	return fmt.Errorf(report)
}
