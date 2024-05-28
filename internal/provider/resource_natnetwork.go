package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	vbg "github.com/mixdone/virtualbox-go"
)

// resourceNatNetwork returns the schema for the NAT network resource.
func resourceNatNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNatNetworkCreate,
		ReadContext:   resourceNatNetworkRead,
		UpdateContext: resourceNatNetworkUpdate,
		DeleteContext: resourceNatNetworkDelete,
		Exists:        resourceNatNetworkExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "NAT Network name.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"network": {
				Description: "The static or DHCP network address and mask of the NAT service interface.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"enabled": {
				Description: "Enabled or disabled the NAT network service.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"dhcp": {
				Description: "Enabled or disabled the DHCP server that you specify by using the 'name' option.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
			},
			"ipv6": {
				Description: "Enabled or disabled IPv6.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"port_forwarding_4": {
				Description: "Enables IPv4 port forwarding by using the rule specified by rule.",
				Type:        schema.TypeList,
				Optional:    true,
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
							Required: true,
						},

						"guestport": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"port_forwarding_6": {
				Description: "Enables IPv6 port forwarding by using the rule specified by rule.",
				Type:        schema.TypeList,
				Optional:    true,
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
							Required: true,
						},

						"guestport": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// resourceNatNetworkCreate creates new NAT network.
func resourceNatNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Retrieving home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	// Initializing VirtualBox client
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	// Creating NAT network configuration
	var natNet vbg.NatNetwork
	natNet.NetName = d.Get("name").(string)
	natNet.Network = d.Get("network").(string)
	natNet.Enabled = d.Get("enabled").(bool)
	natNet.DHCP = d.Get("dhcp").(bool)
	natNet.Ipv6 = d.Get("ipv6").(bool)

	// Retrieving port forwarding rules
	// Creating rules for IPv4 and IPv6
	rules4 := make([]vbg.PortForwarding, 0, 10)
	rules6 := make([]vbg.PortForwarding, 0, 10)
	portForwarding4Number := d.Get("port_forwarding_4.#").(int)
	portForwarding6Number := d.Get("port_forwarding_6.#").(int)

	// Processing IPv4 port forwarding rules
	for i := 0; i < portForwarding4Number; i++ {
		protocol := vbg.TCP
		if d.Get(fmt.Sprintf("port_forwarding_4.%d.protocol", i)).(string) == "udp" {
			protocol = vbg.UDP
		}

		currentPF := vbg.PortForwarding{
			Name:      d.Get(fmt.Sprintf("port_forwarding_4.%d.name", i)).(string),
			Protocol:  protocol,
			HostIP:    d.Get(fmt.Sprintf("port_forwarding_4.%d.hostip", i)).(string),
			HostPort:  d.Get(fmt.Sprintf("port_forwarding_4.%d.hostport", i)).(int),
			GuestIP:   d.Get(fmt.Sprintf("port_forwarding_4.%d.guestip", i)).(string),
			GuestPort: d.Get(fmt.Sprintf("port_forwarding_4.%d.guestport", i)).(int),
		}
		rules4 = append(rules4, currentPF)
	}

	// Processing IPv6 port forwarding rules
	for i := 0; i < portForwarding6Number; i++ {
		protocol := vbg.TCP
		if d.Get(fmt.Sprintf("port_forwarding_6.%d.protocol", i)).(string) == "udp" {
			protocol = vbg.UDP
		}

		currentPF := vbg.PortForwarding{
			Name:      d.Get(fmt.Sprintf("port_forwarding_6.%d.name", i)).(string),
			Protocol:  protocol,
			HostIP:    d.Get(fmt.Sprintf("port_forwarding_6.%d.hostip", i)).(string),
			HostPort:  d.Get(fmt.Sprintf("port_forwarding_6.%d.hostport", i)).(int),
			GuestIP:   d.Get(fmt.Sprintf("port_forwarding_6.%d.guestip", i)).(string),
			GuestPort: d.Get(fmt.Sprintf("port_forwarding_6.%d.guestport", i)).(int),
		}
		rules6 = append(rules6, currentPF)
	}

	// Assigning port forwarding rules to NAT network configuration
	natNet.PortForward4 = rules4
	natNet.PortForward6 = rules6

	// Adding and starting NAT network
	if err := vb.AddNatNet(&natNet); err != nil {
		return diag.Errorf("Adding NAT network failed: %s", err.Error())
	}

	if err := vb.StartNatNet(&natNet); err != nil {
		return diag.Errorf("Starting NAT network failed: %s", err.Error())
	}

	// Setting resource ID to name of NAT network
	d.SetId(natNet.NetName)

	// Triggering read operation to populate resource data
	return resourceNatNetworkRead(ctx, d, m)
}

// resourceNatNetworkRead reads state of existing NAT network.
func resourceNatNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Retrieving home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	// Initializing VirtualBox client
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	// Retrieving list of NAT networks
	natnets, err := vb.ListNatNets()
	if err != nil {
		d.SetId("")
		return diag.Errorf("Getting list of NAT networks failed: %s", err.Error())
	}

	// Finding NAT network matching resource ID
	id := d.Id()

	var necessaryNetwork *vbg.NatNetwork

	for _, i := range natnets {
		if i.NetName == id {
			necessaryNetwork = &i
		}
	}

	// Setting resource data based on retrieved NAT network configuration
	if err := d.Set("name", necessaryNetwork.NetName); err != nil {
		return diag.Errorf("Didn't manage to set name: %s", err.Error())
	}
	if err := d.Set("network", necessaryNetwork.Network); err != nil {
		return diag.Errorf("Didn't manage to set network: %s", err.Error())
	}
	if err := d.Set("enabled", necessaryNetwork.Enabled); err != nil {
		return diag.Errorf("Didn't manage to set enabled or disabled: %s", err.Error())
	}
	if err := d.Set("dhcp", necessaryNetwork.DHCP); err != nil {
		return diag.Errorf("Didn't manage to set enabled or disabled DHCP server: %s", err.Error())
	}
	if err := d.Set("ipv6", necessaryNetwork.Ipv6); err != nil {
		return diag.Errorf("Didn't manage to set enabled or disabled ipv6: %s", err.Error())
	}

	// Populating port forwarding rules for IPv6
	rules4 := make([]map[string]any, 0, 10)
	for i := 0; i < len(necessaryNetwork.PortForward4); i++ {
		protocol := "tcp"
		if necessaryNetwork.PortForward4[i].Protocol == vbg.UDP {
			protocol = "udp"
		}

		rules4 = append(rules4, map[string]any{
			"name":      necessaryNetwork.PortForward4[i].Name,
			"protocol":  protocol,
			"hostip":    necessaryNetwork.PortForward4[i].HostIP,
			"hostport":  necessaryNetwork.PortForward4[i].HostPort,
			"guestip":   necessaryNetwork.PortForward4[i].GuestIP,
			"guestport": necessaryNetwork.PortForward4[i].GuestPort,
		})
	}

	// Setting IPv4 port forwarding rules
	if err := d.Set("port_forwarding_4", rules4); err != nil {
		return diag.Errorf("Didn't manage to set ipv4 port forwarding: %s", err.Error())
	}

	// Populating port forwarding rules for IPv6
	rules6 := make([]map[string]any, 0, 10)
	for i := 0; i < len(necessaryNetwork.PortForward6); i++ {
		protocol := "tcp"
		if necessaryNetwork.PortForward6[i].Protocol == vbg.UDP {
			protocol = "udp"
		}

		rules6 = append(rules6, map[string]any{
			"name":      necessaryNetwork.PortForward6[i].Name,
			"protocol":  protocol,
			"hostip":    necessaryNetwork.PortForward6[i].HostIP,
			"hostport":  necessaryNetwork.PortForward6[i].HostPort,
			"guestip":   necessaryNetwork.PortForward6[i].GuestIP,
			"guestport": necessaryNetwork.PortForward6[i].GuestPort,
		})
	}

	// Setting IPv6 port forwarding rules
	if err := d.Set("port_forwarding_6", rules6); err != nil {
		return diag.Errorf("Didn't manage to set ipv6 port forwarding: %s", err.Error())
	}

	return nil
}

// resourceNatNetworkUpdate updates existing NAT network.
func resourceNatNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Retrieving home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	// Initializing VirtualBox client
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	// Retrieving list of NAT networks
	natnets, err := vb.ListNatNets()
	if err != nil {
		d.SetId("")
		return diag.Errorf("Getting list of NAT networks failed: %s", err.Error())
	}

	// Finding the NAT network matching the resource ID
	id := d.Id()

	var necessaryNetwork *vbg.NatNetwork

	for _, i := range natnets {
		if i.NetName == id {
			necessaryNetwork = &i
		}
	}

	// Collecting parameters for update
	parameters := []string{}

	// Updating network configuration if changed
	actualNetwork := necessaryNetwork.Network
	newNetwork := d.Get("network").(string)
	if actualNetwork != newNetwork {
		parameters = append(parameters, "network")
		necessaryNetwork.Network = newNetwork
	}

	// Updating enabled status if changed
	actualEnabled := necessaryNetwork.Enabled
	newEnabled := d.Get("enabled").(bool)
	if actualEnabled != newEnabled {
		parameters = append(parameters, "enabled")
		necessaryNetwork.Enabled = newEnabled
	}

	// Updating DHCP status if changed
	actualDHCP := necessaryNetwork.DHCP
	newDHCP := d.Get("dhcp").(bool)
	if actualDHCP != newDHCP {
		parameters = append(parameters, "DHCP")
		necessaryNetwork.DHCP = newDHCP
	}

	// Updating IPv6 status if changed
	actualIpv6 := necessaryNetwork.Ipv6
	newIpv6 := d.Get("ipv6").(bool)
	if actualIpv6 != newIpv6 {
		parameters = append(parameters, "ipv6")
		necessaryNetwork.Ipv6 = newIpv6
	}

	// Applying changes if any parameters were updated
	if len(parameters) != 0 {
		err = vb.ModifyNatNet(necessaryNetwork, parameters)
		if err != nil {
			return diag.Errorf("Modify NAT network failed: %s", err.Error())
		}
	}

	// Variable to track if there are changes in port forwarding rules
	needChangeRules := false

	// Slice to hold port forwarding rules to be deleted
	deleteForwardingList := make([]vbg.PortForwarding, 0, 10)

	// Slice to hold new port forwarding rules to be added
	addNewForwardingList := make([]vbg.PortForwarding, 0, 10)

	// Determine number of new IPv4 port forwarding rules
	newRule4Number := d.Get("port_forwarding_4.#").(int)

	// Determine actual number of existing IPv4 port forwarding rules
	actualRule4Number := len(necessaryNetwork.PortForward4)

	// Append empty rules if number of new rules exceeds actual number
	if actualRule4Number < newRule4Number {
		var rules = make([]vbg.PortForwarding, newRule4Number-actualRule4Number)
		necessaryNetwork.PortForward4 = append(necessaryNetwork.PortForward4, rules...)
	}

	// Loop through each new IPv4 port forwarding rule
	for i := 0; i < newRule4Number; i++ {
		// Construct request parameters for current rule
		requestName := fmt.Sprintf("port_forwarding_4.%d.name", i)
		requestHostIp := fmt.Sprintf("port_forwarding_4.%d.hostip", i)
		requestHostPort := fmt.Sprintf("port_forwarding_4.%d.hostport", i)
		requestGuestIp := fmt.Sprintf("port_forwarding_4.%d.guestip", i)
		requestGuestPort := fmt.Sprintf("port_forwarding_4.%d.guestport", i)
		requestProtocol := fmt.Sprintf("port_forwarding_4.%d.protocol", i)

		// Extract values for current rule from Terraform schema
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

		// Check if current rule differs from existing rule
		if currentName != necessaryNetwork.PortForward4[i].Name ||
			currentHostIp != necessaryNetwork.PortForward4[i].HostIP ||
			currentHostPort != necessaryNetwork.PortForward4[i].HostPort ||
			currentGuestIp != necessaryNetwork.PortForward4[i].GuestIP ||
			currentGuestPort != necessaryNetwork.PortForward4[i].GuestPort ||
			currentProtocol != necessaryNetwork.PortForward4[i].Protocol {

			// Mark that changes are needed
			needChangeRules = true

			// Create new port forwarding rule
			rule := vbg.PortForwarding{
				Name:      currentName,
				Protocol:  currentProtocol,
				HostIP:    currentHostIp,
				HostPort:  currentHostPort,
				GuestIP:   currentGuestIp,
				GuestPort: currentGuestPort,
			}

			// Add new rule to list of rules to be added
			addNewForwardingList = append(addNewForwardingList, rule)

			// If current rule index is within actual rule count, add existing rule to list of rules to be deleted
			if i < actualRule4Number {
				deleteForwardingList = append(deleteForwardingList, necessaryNetwork.PortForward4[i])
			}
		}
	}

	// Remove excess IPv4 port forwarding rules if actual count is greater than new count
	if actualRule4Number > newRule4Number {
		for i := 0; i < actualRule4Number-newRule4Number; i++ {
			needChangeRules = true
			deleteForwardingList = append(deleteForwardingList, necessaryNetwork.PortForward4[newRule4Number+i])
		}
	}

	// Apply changes if needed
	if needChangeRules {
		// Delete existing port forwarding rules
		if len(deleteForwardingList) > 0 {
			if err := vb.DeleteAllPortForwNat(necessaryNetwork, deleteForwardingList, "--port-forward-4"); err != nil {
				return diag.Errorf("Unable to delete ipv4 port forwardings: %s", err.Error())
			}
		}

		// Add new port forwarding rules
		if len(addNewForwardingList) > 0 {
			if err := vb.AddAllPortForwNat(necessaryNetwork, addNewForwardingList, "--port-forward-4"); err != nil {
				return diag.Errorf("Unable to set ipv4 port forwardings: %s", err.Error())
			}
		}
	}

	// Reset variables for IPv6 port forwarding rules
	needChangeRules = false
	deleteForwardingList = make([]vbg.PortForwarding, 0, 10)
	addNewForwardingList = make([]vbg.PortForwarding, 0, 10)

	newRule6Number := d.Get("port_forwarding_6.#").(int)
	actualRule6Number := len(necessaryNetwork.PortForward6)

	if actualRule6Number < newRule6Number {
		var rules = make([]vbg.PortForwarding, newRule6Number-actualRule6Number)
		necessaryNetwork.PortForward6 = append(necessaryNetwork.PortForward6, rules...)
	}

	for i := 0; i < newRule6Number; i++ {
		requestName := fmt.Sprintf("port_forwarding_6.%d.name", i)
		requestHostIp := fmt.Sprintf("port_forwarding_6.%d.hostip", i)
		requestHostPort := fmt.Sprintf("port_forwarding_6.%d.hostport", i)
		requestGuestIp := fmt.Sprintf("port_forwarding_6.%d.guestip", i)
		requestGuestPort := fmt.Sprintf("port_forwarding_6.%d.guestport", i)
		requestProtocol := fmt.Sprintf("port_forwarding_6.%d.protocol", i)

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

		if currentName != necessaryNetwork.PortForward6[i].Name ||
			currentHostIp != necessaryNetwork.PortForward6[i].HostIP ||
			currentHostPort != necessaryNetwork.PortForward6[i].HostPort ||
			currentGuestIp != necessaryNetwork.PortForward6[i].GuestIP ||
			currentGuestPort != necessaryNetwork.PortForward6[i].GuestPort ||
			currentProtocol != necessaryNetwork.PortForward6[i].Protocol {

			needChangeRules = true
			rule := vbg.PortForwarding{
				Name:      currentName,
				Protocol:  currentProtocol,
				HostIP:    currentHostIp,
				HostPort:  currentHostPort,
				GuestIP:   currentGuestIp,
				GuestPort: currentGuestPort,
			}
			addNewForwardingList = append(addNewForwardingList, rule)
			if i < actualRule6Number {
				deleteForwardingList = append(deleteForwardingList, necessaryNetwork.PortForward6[i])
			}
		}
	}

	if actualRule6Number > newRule6Number {
		for i := 0; i < actualRule6Number-newRule6Number; i++ {
			needChangeRules = true
			deleteForwardingList = append(deleteForwardingList, necessaryNetwork.PortForward6[newRule4Number+i])
		}
	}

	if needChangeRules {
		if len(deleteForwardingList) > 0 {
			if err := vb.DeleteAllPortForwNat(necessaryNetwork, deleteForwardingList, "--port-forward-6"); err != nil {
				return diag.Errorf("Unable to delete ipv6 port forwardings: %s", err.Error())
			}
		}

		if len(addNewForwardingList) > 0 {
			if err := vb.AddAllPortForwNat(necessaryNetwork, addNewForwardingList, "--port-forward-6"); err != nil {
				return diag.Errorf("Unable to set ipv6 port forwardings: %s", err.Error())
			}
		}
	}

	d.SetId(necessaryNetwork.NetName)

	// Triggering read operation to populate resource data
	return resourceNatNetworkRead(ctx, d, m)
}

// resourceNatNetworkDelete deletes existing NAT network.
func resourceNatNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Retrieving home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return diag.Errorf("userhomedir failed: %s", err.Error())
	}

	// Initializing VirtualBox client
	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	// Retrieving list of NAT networks
	natnets, err := vb.ListNatNets()
	if err != nil {
		d.SetId("")
		return diag.Errorf("Getting list of NAT networks failed: %s", err.Error())
	}

	// Finding NAT network matching the resource ID
	id := d.Id()

	var necessaryNetwork *vbg.NatNetwork

	for _, i := range natnets {
		if i.NetName == id {
			necessaryNetwork = &i
		}
	}

	// Stopping and removing NAT network
	if err := vb.StopNatNet(necessaryNetwork); err != nil {
		return diag.Errorf("Stopping NAT network failed: %s", err.Error())
	}

	if err := vb.RemoveNatNet(necessaryNetwork); err != nil {
		return diag.Errorf("Removing NAT network failed: %s", err.Error())
	}

	return nil
}

func resourceNatNetworkExists(d *schema.ResourceData, m interface{}) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("userhomedir failed: %s", err.Error())
	}

	vb := vbg.NewVBox(vbg.Config{
		BasePath: homeDir,
	})

	natnets, err := vb.ListNatNets()
	if err != nil {
		d.SetId("")
		return false, fmt.Errorf("getting list of NAT networks failed: %s", err)
	}

	id := d.Id()

	for _, i := range natnets {
		if i.NetName == id {
			return true, nil
		}
	}

	return false, nil
}
