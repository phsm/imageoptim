package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

const (
	FALLBACK_PATH = "/var/www"
)

var (
	path         string
	workerscount int
	help         bool
)

func processPNG(p string, wg *sync.WaitGroup, sem chan int) {
	var stderr bytes.Buffer
	cmd := exec.Command("optipng", "-o5", "-quiet", p)
	cmd.Stderr = &stderr
	cmd.Run()
	if stderr.String() != "" {
		fmt.Println(stderr.String())
	}
	wg.Done()
	<-sem
}

func processJPG(p string, wg *sync.WaitGroup, sem chan int) {
	var stderr bytes.Buffer
	cmd := exec.Command("jpegoptim", "--strip-all", "-q", p)
	cmd.Stderr = &stderr
	cmd.Run()
	if stderr.String() != "" {
		fmt.Println(stderr.String())
	}
	wg.Done()
	<-sem
}

func main() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Couldn't get current workdir: %s. That is really strange occasion! Falling back to %s as a workdir.\n", err, FALLBACK_PATH)
		wd = FALLBACK_PATH
	}
	flag.StringVar(&path, "p", wd, "Path to directory containing images e.g. website root. Default: current directory")
	flag.BoolVar(&help, "h", false, "Print help")
	flag.IntVar(&workerscount, "c", 2, "Number of image processing workers")
	flag.Parse()

	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}
	var wg sync.WaitGroup
	sem := make(chan int, workerscount)

	filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
		if strings.HasSuffix(p, ".png") {
			wg.Add(1)
			sem <- 1
			go processPNG(p, &wg, sem)
		}
		if strings.HasSuffix(p, ".jpg") || strings.HasSuffix(p, ".jpe") || strings.HasSuffix(p, ".jpeg") {
			wg.Add(1)
			sem <- 1
			go processJPG(p, &wg, sem)
		}
		return nil
	})
	wg.Wait()
}
