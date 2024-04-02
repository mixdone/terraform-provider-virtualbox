
resource "virtualbox_network" "name" {
  index = 0
}

resource "virtualbox_server" "VM_without_image" {
    count     = 0
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 1000
    status = "running"
    os_id = "Windows7_64"
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
    count     = 0
    name      = format("VM_VDI-%02d", count.index + 1)
    basedir = format("VM_VDI-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
    status = "poweroff"
}



resource "virtualbox_server" "VM_network" {
    count     = 0
    name      = format("VM_network-%02d", count.index + 1)
    basedir = format("VM_network-%02d", count.index + 1)
    cpus      = 3
    memory    = 500

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

    status = "poweroff"
}

# resource "virtualbox_server" "VM_ISO" {
#     count     = 0
#     name      = format("VM_ISO-%02d", count.index + 1)
#     basedir = format("VM_ISO-%02d", count.index + 1)
#     cpus      = 2
#     memory    = 500
#     //image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
# }