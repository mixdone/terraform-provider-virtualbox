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

resource "yandex_compute_instance" "VM1" {
	name = "vm1"
	boot_disk {
		initialize_params {
			image_id = "fd8auu58m9ic4rtekngm"
		}
	}
	network_interface {
		subnet_id = "[192.168.0.0/16]"
		nat = true
	}
	resources {
		cores = 3
		memory = 1
	}
}

resource "yandex_compute_instance" "VM2" {
	name = "vm2"
	boot_disk {
		initialize_params {
			image_id = "fd8auu58m9ic4rtekngm"
		}
	}
	network_interface {
		subnet_id = "[192.168.0.0/16]"
		nat = true
	}
	resources {
		cores = 2
		memory = 2
	}
}

resource "yandex_compute_instance" "VM3" {
	name = "vm3"
	boot_disk {
		initialize_params {
			image_id = "fd8auu58m9ic4rtekngm"
		}
	}
	network_interface {
		subnet_id = "[192.168.0.0/16]"
		nat = true
	}
	resources {
		cores = 1
		memory = 2
	}
}

