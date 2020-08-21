package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

const HostnameUser = "/etc/hostname"
const HostnameKernel = "/proc/sys/kernel/hostname"

const Shell = "/bin/sh"

const InitDir = "/etc/init"
const ZeroDir = InitDir + "/0"
const OneDir = InitDir + "/1"
const OneDirOnce = OneDir + "/once"
const OneDirRepeat = OneDir + "/repeat"
const TwoDir = InitDir + "/2"

func hostname() {
	host, err := ioutil.ReadFile(HostnameUser)
	if err != nil {
		host = []byte("linux")
	}
	ioutil.WriteFile(HostnameKernel, host, 0644)
	fmt.Println("Hostname set.")
}

func RunEachIn(directory string) {
	filepath.Walk(directory, func(current string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		absolute, _ := filepath.Abs(current)
		cmd := exec.Command(Shell, absolute)
		cmd.Run()
		return nil
	})
}

func StartEachIn(directory string) {
	filepath.Walk(directory, func(current string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		absolute, _ := filepath.Abs(current)
		cmd := exec.Command(Shell, absolute)
		go cmd.Run()
		return nil
	})
}

func StartLoopEachIn(directory string) {
	filepath.Walk(directory, func(current string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		absolute, _ := filepath.Abs(current)
		go StartLoop(absolute)
		return nil
	})
}

func StartLoop(absolute string) {
	for {
		cmd := exec.Command(Shell, absolute)
		cmd.Run()
	}
}

func Zero() {
	fmt.Println("-- Phase 0: preliminary system setup.")
	RunEachIn(ZeroDir)
}

func One() {
	fmt.Println("-- Phase 1: launching userspace coroutines.")
	go StartEachIn(OneDirOnce)
	StartLoopEachIn(OneDirRepeat)
	select {}
}

func Two() {
	fmt.Println("-- Phase 2: reboot triggered, shutdown hooks.")
	RunEachIn(TwoDir)
}

func main() {
	if os.Getpid() != 1 {
		fmt.Println("Not run as pid 1.")
		os.Exit(1)
	}
	fmt.Println("Init started.")
	sigc := make(chan os.Signal, 2)
	signal.Notify(sigc, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		sig := <-sigc
		Two()
		switch sig {
		case syscall.SIGUSR1:
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_POWER_OFF)
		case syscall.SIGUSR2:
			syscall.Reboot(syscall.LINUX_REBOOT_CMD_RESTART)
		}
	}()
	go hostname()
	Zero()
	One()
}
