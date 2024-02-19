
resource "virtualbox_server" "VM_without_image" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 500
    status = "poweroff"
    vdi_size = 25000
}

resource "virtualbox_server" "VM_ISO" {
    count     = 0
    name      = format("VM_ISO-%02d", count.index + 1)
    basedir = format("VM_ISO-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    //image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
}