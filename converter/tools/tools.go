// Package tools implements errors, structures and functions which are necessary for the converter to work.
package tools

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var ErrInvalidOperatingSystemName = errors.New("finding suitable yc_os_id failed")
var ErrNoConfigurationFile = errors.New("there is no configuration file for conversion in the directory")
var ErrOsIdNotDefined = errors.New("os_id must be specified")
var ErrCpusNotDefined = errors.New("cpus must be specified")
var ErrMemoryNotDefined = errors.New("memory must be specified")
var ErrInvalidInput = errors.New("invalid input")
var ErrInvalidName = errors.New("invalid resource name")
var ErrInvalidMemoryFormat = errors.New("invalid memory format")
var ErrInvalidCpusFormat = errors.New("invalid cpus format")
var ErrInvalidCountFormat = errors.New("invalid format for number of VMs")

// Info is the structure for the proper credentials.
type Info struct {
	OAuth_token string //IAM (OAuth) token.
	Cloud_ID    string //ID of the cloud to apply any resources to.
	Folder_ID   string //ID of the folder to operate under, if not specified by a given resource.
	Zone        string //The default availability zone to operate under, if not specified by a given resource.
}

// Get_yc_images returns map of the OS images for Yandex Cloud in format map[os_name]os_id and any error encountered.
func Get_yc_images() (map[string]string, error) {
	yc_images := make(map[string]string)

	out, err := os.ReadFile("tools/yc_images.txt")
	if err != nil {
		return nil, err
	}

	splited_yc_images := strings.Split(string(out), "|")

	for i := 7; i < len(splited_yc_images)-2; i += 6 {
		yc_images[strings.ReplaceAll(splited_yc_images[i+1], " ", "")] = strings.ReplaceAll(splited_yc_images[i], " ", "")
	}

	return yc_images, nil
}

// Get_os_name converts the giving name of the OS for VirtualBox to the name of the OS for Yandex Cloud and returns it.
func Get_os_name(vb_os_id string) string {
	os_name := ""

	os_id := strings.Split(vb_os_id, "_")[0]

	for i, ch := range os_id {
		if strings.Contains("0123456789", string(ch)) {
			os_name += "-"
			for j := i; j < len(os_id); j++ {
				if string(os_id[j]) != "\"" {
					os_name += string(os_id[j])
				}
			}
			break
		}
		if string(ch) != "\"" {
			os_name += string(ch)
		}
	}

	os_name += "-"
	os_name = strings.ToLower(os_name)

	return os_name
}

// Get_yc_os_id returns the OS ID for Yandex Cloud by the given OS name or an error if the identifier is not found.
func Get_yc_os_id(yc_images map[string]string, os_name string) (string, error) {
	yc_os_id := ""

	for name, id := range yc_images {
		if strings.Contains(name, os_name) {
			yc_os_id = id
			break
		}
	}

	if yc_os_id == "" {
		return yc_os_id, ErrInvalidOperatingSystemName
	} else {
		return yc_os_id, nil
	}
}

// Choose_option selects 1 option from the giving list of options based on a user choice and returns it and any error encountered.
func Choose_option(options []string) (string, error) {
	var num int
	for i, option := range options {
		fmt.Printf("%d. %s\n", i+1, option)
	}

	if _, err := fmt.Scan(&num); err != nil {
		fmt.Println("Invalid input!")
		return "", ErrInvalidInput
	}

	if (1 <= num) && (num <= 4) {
		return options[num-1], nil
	}

	fmt.Println("Invalid number!")
	return "", ErrInvalidInput
}

// Check_name checks the validity of the resource name for Yandex Сloud.
func Check_name(name string) (bool, error) {
	re := regexp.MustCompile(`^[a-z]+[a-z0-9\-]*[a-z0-9]+$`)
	if !re.MatchString(name) {
		fmt.Printf(`Invalid name "%s"! The name can contain lowercase Latin letters, numbers, and hyphens. 
The first character must be a letter. The last character must not be a hyphen. 
The allowed length is from 2 to 63 characters.`+"\n", name)
		return false, ErrInvalidName
	}
	return true, nil
}
