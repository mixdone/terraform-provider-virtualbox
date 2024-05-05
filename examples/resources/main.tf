
resource "virtualbox_dhcp" "hello" {
  count = 1
  server_ip = "10.0.2.3"
  lower_ip = "10.0.2.23"
  upper_ip = "10.0.2.200"
  network_name = "hohoho"
  network_mask = "255.255.0.0"
  enabled = false
}
