resource "virtualbox_server" "VM_network" {
    count     = 0
    name      = format("VM_network-%02d", count.index + 1)
    basedir = format("VM_network-%02d", count.index + 1)
    cpus      = 3
    memory    = 500
    os_id = "Ubuntu20_64"
    status = "poweroff"
}
