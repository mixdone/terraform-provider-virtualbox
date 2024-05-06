# Output the instance's public IP address.
output "name" {
  value = virtualbox_server.VM_without_image[0].name
}

output "basedir" {
  value = virtualbox_server.VM_without_image[0].basedir
}

output "cpus" {
 value = virtualbox_server.VM_without_image[0].cpus
}

output "memory" {
  value = virtualbox_server.VM_without_image[0].memory
}

output "status" {
  value = virtualbox_server.VM_without_image[0].status
}

output "name_3" {
  value = virtualbox_server.VM_VDI[0].name
}

output "cpus_3" {
  value = virtualbox_server.VM_VDI[0].cpus
}

output "memory_3" {
  value = virtualbox_server.VM_VDI[0].memory
}

output "status_3" {
  value = virtualbox_server.VM_VDI[0].status
}

output "vdi_size_3" {
  value = virtualbox_server.VM_VDI[0].vdi_size
}