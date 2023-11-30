resource "virtualbox_server" "Test3" {
    name = "Test3"
    basedir = "VMS1"
    memory = 300
    url = "https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz"
}
