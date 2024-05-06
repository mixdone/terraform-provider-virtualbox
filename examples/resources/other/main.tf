
# resource "virtualbox_dhcp" "hello" {
#   count = 0
#   server_ip = "10.0.2.3"
#   lower_ip = "10.0.2.23"
#   upper_ip = "10.0.2.200"
#   network_name = "hohoho"
#   network_mask = "255.255.0.0"
#   enabled = false
# }

# resource "virtualbox_server" "VM_without_image" {
#     count     = 1
#     name      = format("VM_without_image-%02d", count.index + 1)
#     basedir = format("VM_without_image-%02d", count.index + 1)
#     cpus      = 3
#     memory    = 1000
#     status = "poweroff"
#     os_id = "Windows7_64"
#     //drag_and_drop = "guesttohost"
#     //clipboard = "guesttohost"

#     network_adapter {
#       network_mode = "nat"
      
#       port_forwarding {
#         name = "lololo"
#         hostport = 63723
#         guestport = 24
#       }
#       port_forwarding {
#         name = "rule2"
#         hostport = 63722
#         guestport = 22
#       }

#     ipv6 = true
#     port_forwarding_6 {
#         name = "rule2"
#         protocol = "udp"
#         hostport = 1022
#         guestip = "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
#         guestport = 21
#     }
# }
# }

# resource "virtualbox_server" "VM_without_image" {
#   count     = 1
#   name      = format("VM_without_image-%02d", count.index + 1)
#   basedir   = format("VM_without_image-%02d", count.index + 1)
#   cpus      = 3
#   memory    = 1000
#   status    = "poweroff"
#   os_id     = "Windows7_64"
#   user_data = file("${path.module}/user_data")
# }

# resource "virtualbox_server" "bad_VM_example" {
#   count   = 0
#   name    = format("VM_without_image-%02d", count.index + 1)
#   basedir = format("VM_without_image-%02d", count.index + 1)
#   cpus    = 3
#   memory  = 2500
#   status  = "poweroff"
#   os_id   = "Windows7_64"
#   group   = "/man"

#   snapshot {
#     name        = "hello"
#     description = "hohohhoho"
#   }
# }


# resource "virtualbox_server" "bad_VM_example" {
#     count     = 1
#     name      = format("VM_without_image-%02d", count.index + 1)
#     basedir = format("VM_without_image-%02d", count.index + 1)
#     cpus      = 30
#     memory    = 1000000000000
#     status = "asdfasdf"
#     os_id = "Windows7_64"
# }

