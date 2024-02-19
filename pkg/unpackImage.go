package pkg

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
)

// download file from url
func FileDownload(url, fpath string) (string, error) {
	file := path.Base(url)
	logrus.Printf("Dowloading file %s from %s\n", file, url)
	var path bytes.Buffer
	path.WriteString(filepath.Join(fpath, file))

	out, err := os.Create(path.String())
	if err != nil {
		return "", fmt.Errorf("creation file failed: %s", err.Error())
	}
	defer out.Close()

	headResp, err := http.Head(url)
	if err != nil {
		return "", fmt.Errorf("http head failed: %s", err.Error())
	}
	defer headResp.Body.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http get failed: %s", err.Error())
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("hopy failed: %s", err.Error())
	}

	logrus.Print("Downloading completed")
	return path.String(), nil
}

// return path to image or virtual disk and error
func UnpackImage(imageArchive, destDir string) (string, error) {
	a, err := os.Open(imageArchive)
	if err != nil {
		return "", fmt.Errorf("open archive failed: %s", err.Error())
	}
	defer a.Close()

	if err = archiver.Unarchive(imageArchive, destDir); err != nil {
		return "", fmt.Errorf("unarchiving failed: %s", err.Error())
	}

	files, err := os.ReadDir(destDir)
	if err != nil {
		return "", fmt.Errorf("read dir failde: %s", err.Error())
	}

	return filepath.Join(destDir, files[0].Name()), nil
}
