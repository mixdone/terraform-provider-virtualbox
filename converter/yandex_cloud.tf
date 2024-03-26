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
	folder_id = "folderId"
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
	folder_id = "folderId"
	service_account_id = "serviceAcc"
	instance_template {
		resources {
			memory = 2
			cores = 2
		}
		boot_disk {
			initialize_params {
				image_id = "fd89ovh4ticpo40dkbvd"
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

resource "yandex_resourcemanager_folder" "folder1" {
	name = "folder1"
}

resource "yandex_compute_instance" "VM2" {
	count = 1
	name = "vm2"
	folder_id = "${yandex_resourcemanager_folder.folder1.id}"
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

resource "yandex_resourcemanager_folder" "folder2" {
	name = "folder2"
}

resource "yandex_compute_instance" "VM3" {
	name = "vm3"
	folder_id = "${yandex_resourcemanager_folder.folder2.id}"
	boot_disk {
		initialize_params {
			image_id = "fd82ubnc7m4sjoi9tssk"
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

resource "yandex_compute_instance" "VM5" {
	count = 2
	name = "vm5"
	folder_id = "${yandex_resourcemanager_folder.folder2.id}"
	boot_disk {
		initialize_params {
			image_id = "fd81v7g3b2g481h03tsp"
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

