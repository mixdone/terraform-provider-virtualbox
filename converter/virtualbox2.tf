resource "virtualbox_server" "VM1" {
    count = 0
    name      = format("VM_without_image-%02d", count.index + 1)
    basedir = format("VM1-%02d", count.index + 1)
    cpus      = 3
    memory    = 500
    status = "running"
    os_id = "Ubuntu20_64"
}
