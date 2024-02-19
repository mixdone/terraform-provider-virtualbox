resource "virtualbox_server" "VM_without_image" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 500

    network_adapter {
        index = 1
        network_mode = "null"
    }
    network_adapter {
        index = 2
        network_mode = "nat"
        nic_type = "82540EM"
        cable_connected = true
    }
    network_adapter {
        index = 3
        network_mode = "hostonly"
    }
    network_adapter {
        index = 4
        network_mode = "bridged"
        nic_type = "virtio"
    }

    status = "poweroff"
}