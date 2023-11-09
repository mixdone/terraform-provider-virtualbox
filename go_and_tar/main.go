package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	file := fileDownload("https://github.com/ccll/terraform-provider-virtualbox-images/releases/download/ubuntu-15.04/ubuntu-15.04.tar.xz", "./")
	if err := unpackImage(file, "./"); err != nil {
		logrus.Fatalf("unpacking failed: %s", err.Error())
	}
}

func fileDownload(url, filepath string) string {
	file := path.Base(url)
	logrus.Printf("Dowloading file %s from %s\n", file, url)
	var path bytes.Buffer
	path.WriteString(filepath + "/" + file)

	out, err := os.Create(path.String())
	if err != nil {
		logrus.Fatalf("Creation file failed: %s", err.Error())
	}
	defer out.Close()

	headResp, err := http.Head(url)
	if err != nil {
		logrus.Fatalf("Http head failed: %s", err.Error())
	}

	fileSize, err := strconv.Atoi(headResp.Header.Get("Content-Length"))
	if err != nil {
		logrus.Fatalf("http get failed: %s", err.Error())

	}

	done := make(chan int64)
	go progressBar(done, int64(fileSize), path.String())

	resp, err := http.Get(url)

	n, err := io.Copy(out, resp.Body)

	done <- n
	fmt.Print(" 100%\n")
	logrus.Print("Downloading completed")
	return file
}

func progressBar(done chan int64, totalSize int64, path string) {
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

func unpackImage(imageArchive, destDir string) error {
	cmd := exec.Command("tar", "-xv", "-C", destDir, "-f", imageArchive)
	return cmd.Run()
}
