
resource "virtualbox_server" "VM_without_image" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 3
    memory    = 1000
    status = "poweroff"
    os_id = "Windows7_64"
}

resource "virtualbox_server" "VM_VDI" {
    count     = 1
    name      = format("VM_VDI-%02d", count.index + 1)
    basedir = format("VM_VDI-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    //url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
    status = "poweroff"
    vdi_size = 25000
}
