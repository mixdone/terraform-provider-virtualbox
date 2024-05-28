
resource "virtualbox_server" "VM_ISO" {
    count     = 0
    name      = format("VM_ISO-%02d", count.index + 1)
    basedir   = format("VM_ISO-%02d", count.index + 1)
    cpus      = 4
    memory    = 2800
    //image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
}

resource "virtualbox_server" "VM_VDI" {
  count   = 0
  name    = format("VM_VDI-%02d", count.index + 1)
  basedir = format("VM_VDI-%02d", count.index + 1)
  cpus    = 2
  memory  = 500
  url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
  status   = "poweroff"
  disk_size = 25000
}

resource "virtualbox_server" "VM_network" {
  count   = 1
  name    = format("VM_network-%02d", count.index + 1)
  basedir = format("VM_network-%02d", count.index + 1)
  cpus    = 3
  memory  = 500

  network_adapter {
    network_mode = "nat"
    port_forwarding {
      name = "rule1"
      hostip = ""
      hostport = "80"
      guestip = ""
      guestport = "63222"
    }
  }
  network_adapter {
    network_mode    = "nat"
    nic_type        = "82540EM"
    cable_connected = true
  }
  network_adapter {
    network_mode = "natnetwork"
    name = "natnet5"
  }
  network_adapter {
    network_mode = "bridged"
    nic_type     = "virtio"
  }

  status = "poweroff"
}

resource "virtualbox_natnetwork" "name" {
  name = "natnet5"
  network = "10.0.0.0/16"
}

resource "virtualbox_dhcp" "name" {
  network_name = "natnet5"
  network_mask = "255.255.0.0"
  server_ip = "10.0.1.2"
  lower_ip = "10.0.0.1"
  upper_ip = "10.0.0.30"
  enabled = true
}

resource "virtualbox_server" "VM_Shapshots" {
    count     = 0
    name      = format("VM_Snapshots-%02d", count.index + 1)
    basedir   = format("VM_Snapshots-%02d", count.index + 1)
    cpus      = 4
    memory    = 2000
    snapshot {
      name = "first"
      description = "example"
      current = true
    }
}
