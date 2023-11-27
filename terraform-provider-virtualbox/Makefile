rebuild: clean macOS_build set_log

macOS_build: set_log
	go build -o terraform-provider-virtualbox
	cp terraform-provider-virtualbox ~/.terraform.d/plugins/terraform-virtualbox.local/virtualboxprovider/virtualbox/1.0.0/darwin_arm64

set_log:
	export TF_LOG=TRACE
	export TF_LOG_PATH="./config/log.txt"

clean:
	rm -rf ./config/.terraform*
	rm -rf ./config/terraform*
	rm -rf ./config/log.txt