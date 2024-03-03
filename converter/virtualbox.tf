resource "virtualbox_server" "VM1" {
    count = 0
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM1-%02d", count.index + 1)
    cpus      = 3
    memory    = 100
    status = "running"
    os_id = "Ubuntu20_64"
    vdi_size = 500
}

resource "virtualbox_server" "VM2" {
    name      = "vm2"
    count = 2
    basedir = format("VM2-%02d", count.index + 1)
    cpus      = 2
    memory    = 2000
    status = "poweroff"
    os_id = "Debian9_64"
}

resource "virtualbox_server" "VM3" {
    name      = "vm3"
    basedir = format("VM3-%02d", count.index + 1)
    cpus      = 1
    memory    = 2000
    os_id = "Fedora_64"
    vdi_size = 1000
}


