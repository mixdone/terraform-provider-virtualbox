
resource "virtualbox_server" "bad_VM_example" {
    count     = 1
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM_without_image-%02d", count.index + 1)
    cpus      = 300
    memory    = 1000000000000
    status = "asdfasdf"
    os_id = "Windows7_64"
}