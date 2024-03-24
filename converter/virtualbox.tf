resource "virtualbox_server" "VM1" {
    count = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM1-%02d", count.index + 1)
    cpus      = 2
    memory    = 100
    status = "running"
    os_id = "Ubuntu20_64"
    vdi_size = 500
    group = "man"
}

resource "virtualbox_server" "VM2" {
    name      = "vm2"
    count = 1
    basedir = format("VM2-%02d", count.index + 1)
    cpus      = 1
    memory    = 2000
    status = "poweroff"
    os_id = "Debian9_64"

    network_adapter {
        network_mode = "nat"
    }
    network_adapter {
        network_mode = "nat"
        nic_type = "82540EM"
        cable_connected = true
    }
    network_adapter {
        network_mode = "hostonly"
    }
    network_adapter {
        network_mode = "bridged"
        nic_type = "virtio"
    }
}

resource "virtualbox_server" "VM3" {
    name      = "vm3"
    basedir = format("VM3-%02d", count.index + 1)
    cpus      = 1
    memory    = 1000
    os_id = "Fedora_64"
    vdi_size = 1000
}

resource "virtualbox_server" "VM4" {
    count = 2
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM1-%02d", count.index + 1)
    cpus      = 1
    memory    = 1000
    status = "running"
    os_id = "Fedora_64"
    vdi_size = 500
    group = "man"

    network_adapter {
        network_mode = "nat"
    }
}


