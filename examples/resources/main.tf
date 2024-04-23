
resource "virtualbox_server" "VM_without_image" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 1000
    status = "poweroff"
    os_id = "Windows7_64"
    //drag_and_drop = "guesttohost"
    //clipboard = "guesttohost"

    network_adapter {
      network_mode = "nat"
      
      port_forwarding {
        name = "lololo"
        hostport = 63723
        guestport = 24
      }
      port_forwarding {
        name = "rule2"
        hostport = 63722
        guestport = 22
      }
      port_forwarding {
        name = "rule3"
        hostport = 63724
        guestport = 21
      }
      port_forwarding {
        name = "rule4"
        hostport = 63726
        guestport = 25
      }
    }
}

resource "virtualbox_server" "bad_VM_example" {
    count     = 0
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 2500
    status = "poweroff"
    os_id = "Windows7_64"
    group = "/man"

    snapshot {
      name = "hello"
      description = "hohohhoho"
      current = true
    }

    snapshot {
      name = "hello2"
      description = "hohohhoho"
      
    }
}


# resource "virtualbox_server" "bad_VM_example" {
#     count     = 1
#     name      = format("VM_without_image-%02d", count.index + 1)
#     basedir = format("VM_without_image-%02d", count.index + 1)
#     cpus      = 30
#     memory    = 1000000000000
#     status = "asdfasdf"
#     os_id = "Windows7_64"
# }

resource "virtualbox_server" "VM_VDI" {
    count     = 1
    name      = format("VM_VDI-%02d", count.index + 1)
    basedir = format("VM_VDI-%02d", count.index + 1)
    cpus      = 20000
    memory    = 500
    //url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
    status = "poweroff"
    vdi_size = 25000
}



resource "virtualbox_server" "VM_network" {
    count     = 0
    name      = format("VM_network-%02d", count.index + 1)
    basedir = format("VM_network-%02d", count.index + 1)
    cpus      = 3000
    memory    = 50000000000

    status = "fsdjalkjflkdsj"
    group = "jalskdfj"

    network_adapter {
        network_mode = "adf"
    }
    network_adapter {
        network_mode = "asdfsadf"
        nic_type = "82jdsjflksjlM"
        cable_connected = true
    }
    network_adapter {
        network_mode = "hostafnly"
    }
    network_adapter {
        network_mode = "bridsadfed"
        nic_type = "virtio"
    }

    snapshot {
      name = "hello"
      description = "eeee"
    }

    snapshot {
      name = "hello2"
      description = "hohoho"
    }
}

# resource "virtualbox_server" "VM_ISO" {
#     count     = 0
#     name      = format("VM_ISO-%02d", count.index + 1)
#     basedir = format("VM_ISO-%02d", count.index + 1)
#     cpus      = 2
#     memory    = 500
#     //image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
# }*/