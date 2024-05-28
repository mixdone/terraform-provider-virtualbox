// Package resources implements functions for creating resources in the configuration file for Yandex Cloud.
package resources

import (
	"converter/tools"
	"fmt"
	"os"
	"slices"
)

// Folder_creation creates "yandex_resourcemanager_folder".
// The function accepts an array with the names of folders that have already been created, the name of the folder being created,
// the configuration file to write to it.
// The function returns updated array with the names of created folders and any error encountered.
func Folder_creation(created_folders []string, folder_name string, new_config *os.File) ([]string, error) {
	if folder_name[0] == '"' && folder_name[len(folder_name)-1] == '"' {
		folder_name = folder_name[1 : len(folder_name)-1]
	}
	if found := slices.Contains(created_folders, folder_name); !found {
		ok, err := tools.Check_name(folder_name)
		if !ok {
			return created_folders, err
		}
		fmt.Fprintf(new_config, "resource \"yandex_resourcemanager_folder\" \"%s\" {\n", folder_name)
		fmt.Fprintf(new_config, "\tname = \"%s\"\n", folder_name)
		fmt.Fprint(new_config, "}\n\n")
		created_folders = append(created_folders, folder_name)
	}
	return created_folders, nil
}
