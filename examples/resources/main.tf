resource "virtualbox_natnetwork" "NatNet1" {
    name = "NatNet1"
    network = "192.168.10.0/24"
    dhcp = false
    port_forwarding_4 {
        name = "rule1"
        protocol = "tcp"
        hostport = 1024
        guestip = "192.168.10.6"
        guestport = 22
    }
    ipv6 = true
    port_forwarding_6 {
        name = "rule2"
        protocol = "udp"
        hostport = 1022
        guestip = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
        guestport = 21
    }
}