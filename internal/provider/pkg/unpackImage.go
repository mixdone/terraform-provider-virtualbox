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
		logrus.Fatalf("Creation file failed: %s", err.Error())
		return "", err
	}
	defer out.Close()

	headResp, err := http.Head(url)
	if err != nil {
		logrus.Fatalf("Http head failed: %s", err.Error())
		return "", err
	}
	defer headResp.Body.Close()

	resp, err := http.Get(url)
	if err != nil {
		logrus.Fatalf("Http get failed: %s", err.Error())
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logrus.Fatalf("Copy failed: %s", err.Error())
		return "", err
	}

	fmt.Print(" 100%\n")
	logrus.Print("Downloading completed")
	return path.String(), nil
}

// return path to image or virtual disk and error
func UnpackImage(imageArchive, destDir string) (string, error) {
	a, err := os.Open(imageArchive)
	if err != nil {
		logrus.Fatalf("Open archive failed: %s", err.Error())
		return "", err
	}
	defer a.Close()

	if err = archiver.Unarchive(imageArchive, destDir); err != nil {
		logrus.Fatalf("Unarchiving failed: %s", err.Error())
		return "", err
	}

	files, err := os.ReadDir(destDir)
	if err != nil {
		logrus.Fatalf("Read dir failde: %s", err.Error())
		return "", err
	}

	return filepath.Join(destDir, files[0].Name()), nil
}
