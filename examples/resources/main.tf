resource "virtualbox_server" "VM_without_image" {
    count     = 0
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 500
    status = "poweroff"
}

resource "virtualbox_server" "VM_VDI" {
    count     = 2
    name      = format("VM_VDI-%02d", count.index + 1)
    basedir = format("VM_VDI-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
    status = "poweroff"
}

resource "virtualbox_server" "VM_ISO" {
    count     = 0
    name      = format("VM_ISO-%02d", count.index + 1)
    basedir = format("VM_ISO-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    //image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
}