terraform {
	required_providers {
		yandex = {
			source = "yandex-cloud/yandex"
		}
	}
}

provider "yandex" {
	token = "token"
	cloud_id = "cloudID"
	folder_id = "folderID"
	zone = "ru-central1-b"
}

resource "yandex_vpc_network" "network-1" {
	name = "network-1"
}

resource "yandex_vpc_subnet" "subnet-1" {
	name = "subnet-1"
	v4_cidr_blocks = ["10.0.0.0/22"]
	zone = "ru-central1-b"
	network_id = "${yandex_vpc_network.network-1.id}"
}

resource "yandex_compute_instance_group" "man" {
	name = "man"
	folder_id = "folderID"
	service_account_id = "serviceAcc"
	instance_template {
		resources {
			memory = 2
			cores = 2
		}
		boot_disk {
			initialize_params {
				image_id = "fd877sidh4gajam1r7vn"
			}
		}
		network_interface {
			network_id = "${yandex_vpc_network.network-1.id}"
			subnet_ids = ["${yandex_vpc_subnet.subnet-1.id}"]
			nat = true
		}
	}
	scale_policy {
		fixed_scale {
			size = 3
		}
	}
	allocation_policy {
		zones = ["ru-central1-b"]
	}
	deploy_policy {
		max_unavailable = 1
		max_expansion = 1
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
		cores = 2
		memory = 2
	}
}

resource "yandex_compute_instance" "VM3" {
	name = "vm3"
	boot_disk {
		initialize_params {
			image_id = "fd8263gk7qeo9om378j1"
		}
	}
	network_interface {
		subnet_id = "${yandex_vpc_subnet.subnet-1.id}"
		nat = true
	}
	resources {
		cores = 2
		memory = 2
	}
}

