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
						"name": {
							Type:     schema.TypeString,
							Required: true,
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
						"port_forwarding": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"protocol": {
										Description: "tcp|udp",
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "tcp",
									},
									"hostip": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
									"hostport": {
										Type:     schema.TypeInt,
										Required: true,
									},
									"guestip": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "",
									},
									"guestport": {
										Type:     schema.TypeInt,
										Required: true,
									},
								},
							},
						},
					},
				},
			},

			"user_data": {
				Description: "Userdata for virtual machine.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
			},

			"os_id": {
				Description: "Specifies the guest OS to run in the VM.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Linux_64",
			},

			"drag_and_drop": {
				Description: "Set drag_and_drop option (disabled | hosttoguest | guesttohost | bidirectional).",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "disabled",
			},

			"clipboard": {
				Description: "Set clipboard option (disabled | hosttoguest | guesttohost | bidirectional).",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "disabled",
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

						"current": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
		},
	}
}

// resourceVirtualBoxCreate creates a virtual machine
// function accepts a ctx context, resource data d, and an interface m representing shared data.
// returns diagnostic messages in case of errors.
func resourceVirtualBoxCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Geting data from config
	if err := validateVmParams(d, true); err != nil {
		return diag.Errorf(err.Error())
	}

	// Initializing structure for storing virtual machine parameters
	var vmConf pkg.VMConfig

	// Getting parameters from ResourceData
	vmConf.Name = d.Get("name").(string)
	vmConf.CPUs = d.Get("cpus").(int)
	vmConf.Memory = d.Get("memory").(int)
	vmConf.DiskSize = int64(d.Get("disk_size").(int))
	vmConf.OS_id = d.Get("os_id").(string)
	vmConf.Group = d.Get("group").(string)
	vmConf.DragAndDrop = d.Get("drag_and_drop").(string)
	vmConf.Clipboard = d.Get("clipboard").(string)

	// Processing snapshots
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
		// Handling the case of an image missing in the configuration
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

			if filepath.Ext(filepath.Base(filename)) == ".vdi" {
				image = filename
			} else if filepath.Ext(filepath.Base(filename)) == ".vhd" {
				image = filename
			} else if filepath.Ext(filepath.Base(filename)) == ".vmdk" {
				image = filename
			} else if filepath.Ext(filepath.Base(filename)) != ".iso" {
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

	rule := make([]vbg.PortForwarding, 0, 10)
	nicNumber := d.Get("network_adapter.#").(int)
	for i := 0; i < nicNumber; i++ {

		requestMode := fmt.Sprintf("network_adapter.%d.network_mode", i)
		currentMode := d.Get(requestMode).(string)

		requestType := fmt.Sprintf("network_adapter.%d.nic_type", i)
		currentType := d.Get(requestType).(string)

		requestCable := fmt.Sprintf("network_adapter.%d.cable_connected", i)
		currentCable := d.Get(requestCable).(bool)

		requestNetworkName := fmt.Sprintf("network_adapter.%d.name", i)
		currentNetworkName := d.Get(requestNetworkName).(string)

		NICs[i].Index = i + 1
		NICs[i].NetworkName = currentNetworkName
		NICs[i].Mode = vbg.NetworkMode(currentMode)
		NICs[i].Type = vbg.NICType(currentType)
		NICs[i].CableConnected = currentCable

		portForwardingNumber := d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.#", i)).(int)

		for j := 0; j < portForwardingNumber; j++ {
			protocol := vbg.TCP
			if d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.protocol", i, j)).(string) == "udp" {
				protocol = vbg.UDP
			}

			currentPF := vbg.PortForwarding{
				NicIndex:  i + 1,
				Index:     i + 1,
				Name:      d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.name", i, j)).(string),
				Protocol:  protocol,
				HostIP:    d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.hostip", i, j)).(string),
				HostPort:  d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.hostport", i, j)).(int),
				GuestIP:   d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.guestip", i, j)).(string),
				GuestPort: d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.%d.guestport", i, j)).(int),
			}
			rule = append(rule, currentPF)
		}
	}

	vmConf.Ltype = ltype
	vmConf.Image_path = image
	vmConf.NICs = NICs[:]

	// Applying network adapter settings to VMConfig
	vmConf.NICs = NICs[:]

	// Creating VM with specified parametrs
	vm, err := pkg.CreateVM(vmConf)
	if err != nil {
		return diag.Errorf("Creation VM failed: %s", err.Error())
	}

	// Setting the VM id for Terraform
	d.SetId(vm.UUIDOrName())

	// Getting information about VM and managing it
	vb := vbg.NewVBox(vbg.Config{BasePath: filepath.Join(homedir, d.Get("basedir").(string))})

	// Updating status of virtual machine
	vm, err = vb.VMInfo(d.Id())
	if err != nil {
		d.SetId("")
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	status := d.Get("status").(string)

	if len(rule) > 0 {
		if err := vb.AddALlPortForw(vm, rule); err != nil {
			return diag.Errorf("Unable to set all port forwardings: %s", err.Error())
		}
	}

	vm.Spec.DragAndDrop = vmConf.DragAndDrop
	vm.Spec.Clipboard = vmConf.Clipboard

	if vmConf.DragAndDrop != "disabled" || vmConf.Clipboard != "disabled" {

		if _, err := vb.ControlVM(vm, "draganddrop"); err != nil {
			return diag.Errorf("Unable to set draganddrop VM: %s", err.Error())
		}

		if _, err := vb.ControlVM(vm, "clipboard mode"); err != nil {
			return diag.Errorf("Unable to set clipboard VM: %s", err.Error())
		}
	}

	if status != "poweroff" {
		if _, err := vb.ControlVM(vm, status); err != nil {
			return diag.Errorf("Unable to set state VM: %s", err.Error())
		}
		vm.Spec.State = vbg.VirtualMachineState(status)
		if err = setState(d, vm); err != nil {
			return diag.Errorf("Setting state failed: %s", err.Error())
		}
	}

	userData := d.Get("user_data").(string)
	if userData != "" {
		vb.SetCloudData("user_data", userData)
	}

	return resourceVirtualBoxRead(ctx, d, m)
}

// resourceVirtualBoxRead reads information about virtual machine
// function accepts a ctx context, resource data d, and an interface m representing shared data.
// returns diagnostic messages in case of errors.
func resourceVirtualBoxRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Getting Machine by id
	homedir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	// Creating basic path for VirtualBox
	basePath := filepath.Join(homedir, d.Get("basedir").(string))
	vb := vbg.NewVBox(vbg.Config{BasePath: basePath})
	vm, err := vb.VMInfo(d.Id())

	if err != nil {
		d.SetId("")
		return diag.Errorf("VMInfo failed: %s", err.Error())
	}

	val := d.Get("user_data").(string)
	if val != "" {
		userData, err := vb.GetCloudData("user_data")
		if err != nil {
			return diag.Errorf("Failed to get cloud-config: %v", err.Error())
		}
		if userData != nil && *userData != "" {
			err = d.Set("user_data", *userData)
			if err != nil {
				return diag.Errorf("Failed to set cloud-config: %v", err.Error())
			}
		}
	}

	if err := d.Set("drag_and_drop", vm.Spec.DragAndDrop); err != nil {
		return diag.Errorf("Didn't manage to set drag and drop: %s", err.Error())
	}

	if err := d.Set("clipboard", vm.Spec.Clipboard); err != nil {
		return diag.Errorf("Didn't manage to set clipboard: %s", err.Error())
	}

	// Set state of Machine for Terraform
	if err := setState(d, vm); err != nil {
		return diag.Errorf("Didn't manage to set VMState: %s", err.Error())
	}

	// Set name of Machine for Terraform
	if err := d.Set("name", vm.Spec.Name); err != nil {
		return diag.Errorf("Didn't manage to set name: %s", err.Error())
	}

	// Set CPUs amount for Terraform
	if err := d.Set("cpus", vm.Spec.CPU.Count); err != nil {
		return diag.Errorf("Didn't manage to set cpus: %s", err.Error())
	}

	// Set memory for Terraform
	if err := d.Set("memory", vm.Spec.Memory.SizeMB); err != nil {
		return diag.Errorf("Didn't manage to set memory: %s", err.Error())
	}

	// Set network for Terraform
	if err := setNetwork(d, vm); err != nil {
		return diag.Errorf("Didn't manage to set Network: %s", err.Error())
	}

	// Set snapshots for Terraform
	if err := setSnapshots(d, vm); err != nil {
		return diag.Errorf("Didn't manage to set snapshots: %s", err.Error())
	}

	// Set basedir VM for Terraform
	if err := d.Set("basedir", d.Get("basedir").(string)); err != nil {
		return diag.Errorf("Didn't manage to set basedir: %s", err.Error())
	}

	return nil
}

// poweroffVM performs shutdown of virtual machine
// function accepts a ctx context, resource data d, and an interface m representing shared data.
// returns diagnostic messages in case of errors.
func poweroffVM(vm *vbg.VirtualMachine, vb *vbg.VBox) error {
	// Checking current state of virtual machine
	switch vm.Spec.State {
	case vbg.Poweroff, vbg.Aborted, vbg.Saved:
		return nil
	}

	// Shutting down virtual machine
	if _, err := vb.ControlVM(vm, "poweroff"); err != nil {
		logrus.Errorf("Unable to poweroff VM: %s", err.Error())
		return err
	}

	// Setting virtual machine status to "poweroff"
	vm.Spec.State = vbg.Poweroff
	return nil
}

// resourceVirtualBoxUpdate updates virtual machine settings
// function accepts a ctx context, resource data d, and an interface m representing shared data.
// returns diagnostic messages in case of errors.
func resourceVirtualBoxUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	// Checking parameters of virtual machine
	if err := validateVmParams(d, false); err != nil {
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
	if err = poweroffVM(vm, vb); err != nil {
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
	needChangeRules := false
	deleteForwardingList := make([]vbg.PortForwarding, 0, 10)
	addNewForwardingList := make([]vbg.PortForwarding, 0, 10)

	nicNumber := d.Get("network_adapter.#").(int)

	// Adding new network adapters if necessary
	if len(vm.Spec.NICs) < nicNumber {
		var NICs = make([]vbg.NIC, nicNumber-len(vm.Spec.NICs))
		vm.Spec.NICs = append(vm.Spec.NICs, NICs...)
	}

	// Iterating through and updating parameters of each network adapter
	for i := 0; i < nicNumber; i++ {
		vm.Spec.NICs[i].Index = i + 1

		// Updating operating mode of network adapter
		requestMode := fmt.Sprintf("network_adapter.%d.network_mode", i)
		currentMode := vbg.NetworkMode(d.Get(requestMode).(string))
		if currentMode != vm.Spec.NICs[i].Mode {
			needAppendNetwork = true
			vm.Spec.NICs[i].Mode = currentMode
		}

    // Updating name of network adapter
		requestNetworkName := fmt.Sprintf("network_adapter.%d.name", i)
		currentNetworkName := d.Get(requestNetworkName).(string)
		if currentNetworkName != vm.Spec.NICs[i].NetworkName {
			needAppendNetwork = true
			vm.Spec.NICs[i].NetworkName = currentNetworkName
		}

		// Updating type of network adapter

		requestType := fmt.Sprintf("network_adapter.%d.nic_type", i)
		currentType := vbg.NICType(d.Get(requestType).(string))
		if currentType != vm.Spec.NICs[i].Type {
			needAppendNetwork = true
			vm.Spec.NICs[i].Type = currentType
		}

		// Updating connection status of network adapter cable
		requestCable := fmt.Sprintf("network_adapter.%d.cable_connected", i)
		currentCable := d.Get(requestCable).(bool)
		if currentCable != vm.Spec.NICs[i].CableConnected {
			needAppendNetwork = true
			vm.Spec.NICs[i].CableConnected = currentCable
		}

		ruleNumber := d.Get(fmt.Sprintf("network_adapter.%d.port_forwarding.#", i)).(int)
		amountOfRulesInVM := len(vm.Spec.NICs[i].PortForwarding)
		if len(vm.Spec.NICs[i].PortForwarding) < ruleNumber {
			var rules = make([]vbg.PortForwarding, ruleNumber-len(vm.Spec.NICs[i].PortForwarding))
			vm.Spec.NICs[i].PortForwarding = append(vm.Spec.NICs[i].PortForwarding, rules...)
		}

		nic := vm.Spec.NICs[i]

		for j := 0; j < ruleNumber; j++ {
			requestName := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.name", i, j)
			requestHostIp := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.hostip", i, j)
			requestHostPort := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.hostport", i, j)
			requestGuestIp := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.guestip", i, j)
			requestGuestPort := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.guestport", i, j)
			requestProtocol := fmt.Sprintf("network_adapter.%d.port_forwarding.%d.protocol", i, j)

			currentName := d.Get(requestName).(string)
			currentHostIp := d.Get(requestHostIp).(string)
			currentHostPort := d.Get(requestHostPort).(int)
			currentGuestIp := d.Get(requestGuestIp).(string)
			currentGuestPort := d.Get(requestGuestPort).(int)
			val := d.Get(requestProtocol).(string)
			currentProtocol := vbg.TCP
			if val == "udp" {
				currentProtocol = vbg.UDP
			}

			if currentName != nic.PortForwarding[j].Name ||
				currentHostIp != nic.PortForwarding[j].HostIP ||
				currentHostPort != nic.PortForwarding[j].HostPort ||
				currentGuestIp != nic.PortForwarding[j].GuestIP ||
				currentGuestPort != nic.PortForwarding[j].GuestPort ||
				currentProtocol != nic.PortForwarding[j].Protocol {

				needChangeRules = true
				rule := vbg.PortForwarding{
					Index:     j + 1,
					NicIndex:  i + 1,
					Name:      currentName,
					Protocol:  currentProtocol,
					HostIP:    currentHostIp,
					HostPort:  currentHostPort,
					GuestIP:   currentGuestIp,
					GuestPort: currentGuestPort,
				}
				addNewForwardingList = append(addNewForwardingList, rule)
				if j < amountOfRulesInVM {
					deleteForwardingList = append(deleteForwardingList, nic.PortForwarding[j])
				}
			}
		}

		if len(nic.PortForwarding) > ruleNumber {
			ln := len(nic.PortForwarding)
			for j := 0; j < ln-ruleNumber; j++ {
				needChangeRules = true
				deleteForwardingList = append(deleteForwardingList, nic.PortForwarding[ln+j-1])
			}
		}
	}

	if needAppendNetwork {
		parameters = append(parameters, "network_adapter")
	}

	// Updating VM group
	group := d.Get("group").(string)
	if vm.Spec.Group != group {
		parameters = append(parameters, "group")
		vm.Spec.Group = group
	}

	dragAndDrop := d.Get("drag_and_drop").(string)
	if vm.Spec.DragAndDrop != dragAndDrop {
		parameters = append(parameters, "drag_and_drop")
		vm.Spec.DragAndDrop = dragAndDrop
	}

	clipboardMode := d.Get("clipboard").(string)
	if vm.Spec.Clipboard != clipboardMode {
		parameters = append(parameters, "clipboard")
		vm.Spec.Clipboard = clipboardMode
	}

	// Modify VM
	if len(parameters) != 0 {
		err = vb.ModifyVM(vm, parameters)
		if err != nil {
			return diag.Errorf("ModifyVM failed: %s", err.Error())
		}
	}

	if needChangeRules {
		if len(deleteForwardingList) > 0 {
			if err := vb.DeleteAllPortForw(vm, deleteForwardingList); err != nil {
				return diag.Errorf("Unable to delete port forwardings: %s", err.Error())
			}
		}

		if len(addNewForwardingList) > 0 {
			if err := vb.AddALlPortForw(vm, addNewForwardingList); err != nil {
				return diag.Errorf("Unable to set port forwardings: %s", err.Error())
			}
		}
	}

	// Setting new VM id
	vm.UUID = vm.UUIDOrName()
	d.SetId(vm.UUIDOrName())

	// Updating state
	status := d.Get("status").(string)

	// Virtual machine status management (startup/shutdown)
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

	// Updating Virtual Machine snapshots
	snapshots := d.Get("snapshot.#").(int)

	diff := len(vm.Spec.Snapshots) - snapshots

	if diff < 0 {
		emptySnapshts := make([]vbg.Snapshot, -diff)
		vm.Spec.Snapshots = append(vm.Spec.Snapshots, emptySnapshts...)
	}

	var currentSnap vbg.Snapshot

	// Processing adding/updating/deleting snapshots
	for i := 0; i < snapshots; i++ {
		var snapshot vbg.Snapshot
		req1 := fmt.Sprintf("snapshot.%d.name", i)
		req2 := fmt.Sprintf("snapshot.%d.description", i)
		req3 := fmt.Sprintf("snapshot.%d.current", i)

		snapshot.Name = d.Get(req1).(string)
		snapshot.Description = d.Get(req2).(string)
		current := d.Get(req3).(bool)
		if current {
			currentSnap = snapshot
		}

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

	if currentSnap.Name != "" {
		if err := vb.RestoreSnapshot(vm, currentSnap); err != nil {
			return diag.Errorf("Snapshot restore failed: %s", err.Error())
		}
	}

	return resourceVirtualBoxRead(ctx, d, m)
}

// snapshotOperationsHandler processes virtual machine snapshot operations
// function accepts a pointer to VirtualBox vb object, a pointer to VirtualMachine vm object,
// previous prevSnapshot snapshot, current snapshot, operation type and status
// returns an error if operation failed
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

// resourceVirtualBoxExists checks existence of VirtualBox VM by its ID
// function accepts a pointer to schema object.ResourceData d, which contains virtual machine ID,
// and interface m, which represents execution context
// returns a boolean value
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

// resourceVirtualBoxDelete deletes virtual machine
// function accepts a ctx context, resource data d, and an interface m representing shared data.
// returns diagnostic messages in case of errors.
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
	if err = poweroffVM(vm, vb); err != nil {
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

// setState sets state of virtual machine in schema object.ResourceData
// function accepts a pointer to schema object.ResourceData d, which represents state of resource,
// and a pointer to vbs.VirtualMachine vm object, which contains information about state of virtual machine
// function sets value "status" in object d according to current state of vm
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

func setSnapshots(d *schema.ResourceData, vm *vbg.VirtualMachine) error {
	arr := make([]map[string]interface{}, 0, 3)

	for i := 0; i < len(vm.Spec.Snapshots); i++ {
		if vm.Spec.Snapshots[i].Name == vm.Spec.CurrentSnapshot.Name &&
			vm.Spec.Snapshots[i].Description == vm.Spec.CurrentSnapshot.Description {
			arr = append(arr, map[string]interface{}{
				"name":        vm.Spec.Snapshots[i].Name,
				"description": vm.Spec.Snapshots[i].Description,
				"current":     true,
			})
		} else {
			arr = append(arr, map[string]interface{}{
				"name":        vm.Spec.Snapshots[i].Name,
				"description": vm.Spec.Snapshots[i].Description,
				"current":     false,
			})
		}

	}

	if err := d.Set("snapshot", arr); err != nil {
		return fmt.Errorf("%s", err.Error())
	}
	return nil
}

// setNetwork sets information about network adapters in schema object.ResourceData
// based on data about network interfaces of virtual machine
// function accepts a pointer to schema object.ResourceData d, which represents state of resource,
// and a pointer to vbs.VirtualMachine vm object, which contains information about network interfaces of VM
// function creates an array with information about each network adapter of virtual machine and
// installs it in object d under key "network_adapter", each element of array contains adapter index,
// network mode, adapter type, and cable connection status
func setNetwork(d *schema.ResourceData, vm *vbg.VirtualMachine) error {

	// getType helper function returns a string representation of type of network adapter
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

	// getMode helper function returns a string representation of network mode for network adapter
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

	//velociped
	if len(vm.Spec.NICs) == 1 {
		if vm.Spec.NICs[0].Mode == "nat" && vm.Spec.NICs[0].Type == "82540EM" {
			return nil
		}
	}

	// Creating empty array to store information about network adapters
	nics := make([]map[string]any, 0, 4)
	// Iterating through all network adapters of virtual machine and create information about each adapter
	for i, nic := range vm.Spec.NICs {
		out := make(map[string]any)
		out["index"] = i + 1
		out["network_mode"] = getMode(nic)
		out["nic_type"] = getType(nic)
		out["cable_connected"] = nic.CableConnected
		out["name"] = nic.NetworkName

		rules := make([]map[string]any, 0, 3)
		for j := 0; j < len(nic.PortForwarding); j++ {
			protocol := "tcp"
			if nic.PortForwarding[j].Protocol == vbg.UDP {
				protocol = "udp"
			}

			rules = append(rules, map[string]any{
				"name":      nic.PortForwarding[j].Name,
				"protocol":  protocol,
				"hostip":    nic.PortForwarding[j].HostIP,
				"hostport":  nic.PortForwarding[j].HostPort,
				"guestip":   nic.PortForwarding[j].GuestIP,
				"guestport": nic.PortForwarding[j].GuestPort,
			})
		}
		out["port_forwarding"] = rules

		nics = append(nics, out)
	}

	if err := d.Set("network_adapter", nics); err != nil {
		return err
	}

	return nil
}

// validateVmParams checks VM parameters passed in schema object.ResourceData d for correctness
// function returns an error if problems with parameters are detected
// parameters to be checked:
// - number of processors: must be greater than 0 and not exceed number of available processors in system
// - memory size: must be greater than 0 and not exceed total amount of available memory in system
// - status: must match one of the following values: "poweroff", "running", "paused", "saved", "aborted"
func validateVmParams(d *schema.ResourceData, isCreate bool) error {
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

	val, ok := d.GetOk("group")
	if ok {
		group := val.(string)
		if group != "" && group[0] != '/' && group[0] != '\\' {
			error_output = append(error_output, fmt.Sprintf("Not a path in group field, try /%v", group))
			amountOfProblems++
		}
	}

	modes := [4]string{"disabled", "hosttoguest", "guesttohost", "bidirectional"}

	dragAndDrop := d.Get("drag_and_drop").(string)
	badformat := true
	for id := range modes {
		if modes[id] == dragAndDrop {
			badformat = false
			break
		}
	}

	if badformat {
		error_output = append(error_output, "Invalid drag_and_drop option, check description.")
		amountOfProblems++
	}

	badformat = true
	clipboard := d.Get("clipboard").(string)
	for id := range modes {
		if modes[id] == clipboard {
			badformat = false
			break
		}
	}

	if badformat {
		error_output = append(error_output, "Invalid clipboard option, check description.")
		amountOfProblems++
	}

	snapshots := d.Get("snapshot.#").(int)
	if isCreate && snapshots > 1 {
		error_output = append(error_output, "Too many snapshots for a new VM")
		amountOfProblems++
	}

	counter := 0
	for i := 0; i < snapshots; i++ {
		req1 := fmt.Sprintf("snapshot.%d.current", i)
		if d.Get(req1).(bool) {
			counter++
		}
	}
	if counter > 1 {
		error_output = append(error_output, "Only one snapshot can be current")
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
		error_output = append(error_output, "Status does not match any of the existing ones\n\t- poweroff\n\t- runnning\n\t- paused \n\t- saved \n\t- aborted")
		amountOfProblems++
	}

	checkMode := func(nicMode string) error {
		switch nicMode {
		case "none":
			return nil
		case "null":
			return nil
		case "nat":
			return nil
		case "natnetwork":
			return nil
		case "bridged":
			return nil
		case "intnet":
			return nil
		case "hostonly":
			return nil
		case "generic":
			return nil
		default:
			return fmt.Errorf("mode does not match any of the existing ones")
		}
	}
	checkType := func(nicType string) error {
		switch nicType {
		case "Am79C970A":
			return nil
		case "Am79C973":
			return nil
		case "82540EM":
			return nil
		case "82543GC":
			return nil
		case "82545EM":
			return nil
		case "virtio":
			return nil
		default:
			return fmt.Errorf("type does not match any of the existing ones")
		}
	}

	amountOfNICs := d.Get("network_adapter.#").(int)
	var err error
	badNICsMode := false
	badNICsType := false
	allNICReports := "\n"

	for i := 0; i < amountOfNICs; i++ {
		nicReport := fmt.Sprintf("  NIC %d:\n", i)
		badFormat := false

		requestMode := fmt.Sprintf("network_adapter.%d.network_mode", i)
		currentMode := d.Get(requestMode).(string)
		if err = checkMode(currentMode); err != nil {
			nicReport += fmt.Sprintf("\t%s\n", err)
			badFormat = true
			badNICsMode = true
		}

		requestType := fmt.Sprintf("network_adapter.%d.nic_type", i)
		currentType := d.Get(requestType).(string)
		if err = checkType(currentType); err != nil {
			nicReport += fmt.Sprintf("\t%s\n", err)
			badFormat = true
			badNICsType = true
		}

		if badFormat {
			allNICReports += nicReport + "\n"
		}
	}

	if badNICsMode || badNICsType {
		if badNICsMode {
			allNICReports += "\n  NIC modes:\n\t- none\n\t- null\n\t- nat\n\t- natnetwork\n\t- bridget\n\t- intnet\n\t- hostonly\n\t- generic\n"
		}

		if badNICsType {
			allNICReports += "\n  NIC types:\n\t- Am79C970A\n\t- Am79C970A\n\t- 82540EM\n\t- 82543GC\n\t- 82545EM\n\t- virtio\n"
		}
		amountOfProblems++
		error_output = append(error_output, allNICReports)
	}

	if amountOfProblems == 0 {
		return nil
	}

	report := fmt.Sprintf("VM: %v\n", d.Get("name").(string))
	for i := 1; i <= amountOfProblems; i++ {
		report += fmt.Sprintf("%v) %s\n", i, error_output[i-1])
	}
	return fmt.Errorf(report)
}
