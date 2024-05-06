
resource "virtualbox_server" "bad_VM_example1" {
    count     = 1
    name      = format("VM_bad_example1-%02d", count.index + 1)
    basedir = format("VM_bad_example1-%02d", count.index + 1)
    cpus      = 3000
    memory    = 1000000000000
    group = "ajdflkj"
    status = "asdfasdf"
    os_id = "Windows7_64"
}

resource "virtualbox_server" "bad_VM_example2" {
    count     = 1
    name      = format("VM_bad_example_2-%02d", count.index + 1)
    basedir = format("VM_bad_example_2-%02d", count.index + 1)
    cpus      = 3000
    memory    = 1000000000000
    group = "ajdflkj"
    status = "asdfasdf"
    os_id = "Windows7_64"

    network_adapter {
        network_mode = "sadfl"
    }
    
    network_adapter {
        network_mode    = "nat"
        nic_type        = "82540EM"
        cable_connected = true
    }
    network_adapter {
        network_mode = "hostonlyy"
    }
    network_adapter {
        network_mode = "bridged"
        nic_type     = "virtiosdaf"
    }
}