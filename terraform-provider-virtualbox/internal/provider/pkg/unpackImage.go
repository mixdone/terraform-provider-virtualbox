package pkg

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/mholt/archiver"
	"github.com/sirupsen/logrus"
)

func FileDownload(url, filepath string) (string, error) {
	file := path.Base(url)
	logrus.Printf("Dowloading file %s from %s\n", file, url)
	var path bytes.Buffer
	path.WriteString(filepath + "/" + file)

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

	fileSize, err := strconv.Atoi(headResp.Header.Get("Content-Length"))
	if err != nil {
		logrus.Fatalf("Strconv failed: %s", err.Error())
		return "", err
	}

	done := make(chan int64)
	go ProgressBar(done, int64(fileSize), path.String())

	resp, err := http.Get(url)
	if err != nil {
		logrus.Fatalf("Http get failed: %s", err.Error())
		return "", err
	}

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		logrus.Fatalf("Copy failed: %s", err.Error())
		return "", err
	}

	done <- n
	fmt.Print(" 100%\n")
	logrus.Print("Downloading completed")
	return filepath + file, nil
}

func ProgressBar(done chan int64, totalSize int64, path string) {
	file, err := os.Open(path)
	if err != nil {
		logrus.Fatalf("File open filaed: %s", err.Error())
	}
	defer file.Close()
	percent := 0.0

	var stop bool = false
	for {
		select {
		case <-done:
			stop = true
		default:
			fi, err := file.Stat()
			if err != nil {
				log.Fatal(err)
			}

			size := fi.Size()
			if size == 0 {
				size = 1
			}

			var new_percent float64 = float64(size) / float64(totalSize) * 100

			for new_percent > percent {
				fmt.Print("█")
				percent++
			}

		}
		if stop {
			break
		}
		time.Sleep(time.Second)
	}
}

func UnpackImage(imageArchive, destDir string) error {
	a, err := os.Open(imageArchive)
	if err != nil {
		logrus.Fatalf("Open archive failed: %s", err.Error())
		return err
	}
	defer a.Close()

	return archiver.Unarchive(imageArchive, destDir)
}
