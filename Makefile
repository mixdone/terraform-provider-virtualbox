rebuild: clean macOS_build

macOS_build: set_log
	go build -o terraform-provider-virtualbox
	cp terraform-provider-virtualbox ~/.terraform.d/plugins/terraform-virtualbox.local/virtualboxprovider/virtualbox/1.0.0/darwin_arm64

windows_386_build: set_log
	go build -o terraform-provider-virtualbox.exe
	cp terraform-provider-virtualbox.exe  ~\AppData\Roaming\terraform.d\plugins\terraform-virtualbox.local\virtualboxprovider\virtualbox\1.0.0\windows_386

linux_amd64_build: set_log
	go build -o terraform-provider-virtualbox
	cp terraform-provider-virtualbox ~/.terraform.d/plugins/terraform-virtualbox.local/virtualboxprovider/virtualbox/1.0.0/linux_amd64

set_log:
	export TF_LOG=TRACE
	export TF_LOG_PATH="./examples/resources/log.txt"

clean:
	rm -rf ./examples/resources/.terraform*
	rm -rf ./examples/resources/terraform*
	rm -rf ./examples/resources/log.txt
