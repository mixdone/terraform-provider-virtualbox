// Package resources implements functions for creating resources in the configuration file for Yandex Cloud.
package resources

import (
	"converter/tools"
	"fmt"
	"net/netip"
	"os"
	"strconv"
	"strings"
)

// Network_creation creates "yandex_vpc_network".
// The function accepts the configuration file to write to it and returns the name of the created network and any error encountered.
func Network_creation(new_config *os.File) (string, error) {
	var network_name string
	for {
		fmt.Println("Please write a name for the network.")
		if _, err := fmt.Scan(&network_name); err != nil {
			return "", err
		}

		if ok, _ := tools.Check_name(network_name); ok {
			break
		}
	}

	fmt.Fprintf(new_config, "resource \"yandex_vpc_network\" \"%s\" {\n", network_name)
	fmt.Fprintf(new_config, "\tname = \"%s\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return network_name, nil
}

// Subnet_creation creates "yandex_vpc_subnet".
// The function accepts the configuration file to write to it, the name of the network this subnet belongs to, name of the Yandex Cloud zone for this subnet.
// The function returns the name of the created subnet and any error encountered.
func Subnet_creation(new_config *os.File, network_name string, zone string) (string, error) {
	var subnet_name string
	for {
		fmt.Println("Please write a name for the subnet.")
		_, err := fmt.Scan(&subnet_name)
		if err != nil {
			return "", err
		}

		if ok, _ := tools.Check_name(subnet_name); ok {
			break
		}
	}

	fmt.Fprintf(new_config, "resource \"yandex_vpc_subnet\" \"%s\" {\n", subnet_name)
	fmt.Fprintf(new_config, "\tname = \"%s\"\n", subnet_name)

	var v4_cidr_blocks string
	for {
		fmt.Println(`Please write a list of blocks of internal IPv4 addresses that are owned by this subnet.
For example, 10.0.0.0/22 or 192.168.0.0/16.
Blocks of addresses must be unique and non-overlapping within a network.
Minimum subnet size is /28, and maximum subnet size is /16. Only IPv4 is supported.`)
		_, err := fmt.Scan(&v4_cidr_blocks)
		if err != nil {
			return "", err
		}
		splited := strings.Split(v4_cidr_blocks, "/")
		if len(splited) < 2 {
			fmt.Println("Invalid input!")
			continue
		}
		mask, err := strconv.Atoi(splited[1])
		if (err != nil) || (16 > mask) || (mask > 28) {
			fmt.Println("Invalid subnet size!")
			continue
		}
		_, err = netip.ParseAddr(splited[0])
		if err != nil {
			fmt.Println("Invalid IPv4 addresses!")
			continue
		}
		break
	}
	fmt.Fprintf(new_config, "\tv4_cidr_blocks = [\"%s\"]\n", v4_cidr_blocks)

	fmt.Fprintf(new_config, "\tzone = \"%s\"\n", zone)
	fmt.Fprintf(new_config, "\tnetwork_id = \"${yandex_vpc_network.%s.id}\"\n", network_name)
	fmt.Fprint(new_config, "}\n\n")
	return subnet_name, nil
}
