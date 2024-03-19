terraform {
  required_providers {
    yandex = {
      source = "yandex-cloud/yandex"
    }
  }
}

provider "yandex" {
  token     = "<OAuth_token>"
  cloud_id  = "<cloud_ID>"
  folder_id = "<folder_ID>"
  zone      = "ru-central1-a"
}

resource "yandex_vpc_network" "network-1" {
	name =  "network-1"
}

resource "yandex_vpc_subnet" "subnet-1" {
	name =  "subnet-1"
	v4_cidr_blocks = ["10.0.0.0/28"]
	zone = "ru-central1-a"
	network_id = "${yandex_vpc_network.network-1.id}"
}

resource "yandex_compute_instance" "VM1" {
	count = 0
	name = format("VM_without_image-%02d", count.index + 1)
	boot_disk {
		initialize_params {
			image_id = "fd80jfslq61mssea4ejn"
		}
	}
	network_interface {
		subnet_id = "${yandex_vpc_subnet.subnet-1.id}"
		nat = true
	}
	resources {
		cores = 3
		memory = 1
	}
}

resource "yandex_compute_instance" "VM2" {
	count = 1
	name = "vm2"
	boot_disk {
		initialize_params {
			image_id = "fd833q45aucu0afdc2vj"
		}
	}
	network_interface {
		subnet_id = "${yandex_vpc_subnet.subnet-1.id}"
		nat = true
	}
	resources {
		cores = 1
		memory = 2
	}
}

resource "yandex_compute_instance" "VM3" {
	name = "vm3"
	boot_disk {
		initialize_params {
			image_id = "fd898bkh38ssva3kb5td"
		}
	}
	network_interface {
		subnet_id = "${yandex_vpc_subnet.subnet-1.id}"
		nat = true
	}
	resources {
		cores = 1
		memory = 1
	}
}

