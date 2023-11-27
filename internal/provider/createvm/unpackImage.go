package createvm

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

	resp, _ := http.Get(url)

	n, _ := io.Copy(out, resp.Body)

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
	a, err := os.Open(imageArchive)
	if err != nil {
		logrus.Fatalf("Open archive failed: %s", err.Error())
		return err
	}
	defer a.Close()

	return archiver.Unarchive(imageArchive, destDir)
}
