resource "virtualbox_server" "VM_without_image" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
}

resource "virtualbox_server" "VM_VDI" {
    count     = 0
    name      = format("VM_VDI-%02d", count.index + 1)
    basedir = format("VM_VDI-%02d", count.index + 1)
    cpus      = 2
    memory    = 500
    url =  "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
}

# resource "virtualbox_server" "VM" {
#     count     = 0
#     name      = format("VM-%02d", count.index + 1)
#     basedir = format("VM-%02d", count.index + 1)
#     cpus      = 2
#     memory    = 500
#     image = "C:/Users/vovap/ubuntu-16.04.6-desktop-i386.iso"
# }