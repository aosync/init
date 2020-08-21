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
const EntropyReserve = "/var/entropy"
const EntropyKernel = "/dev/random"

const Shell = "/bin/sh"

const InitDir = "/etc/init"
const ZeroDir = InitDir + "/0"
const OneDir = InitDir + "/1"
const OneDirOnce = OneDir + "/once"
const OneDirRepeat = OneDir + "/repeat"
const TwoDir = InitDir + "/2"

func system(prog string, args ...string) {
	cmd := exec.Command(prog, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
}

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
		system(Shell, absolute)
		return nil
	})
}

func StartEachIn(directory string) {
	filepath.Walk(directory, func(current string, info os.FileInfo, err error) error {
		if info.IsDir() || err != nil {
			return nil
		}
		absolute, _ := filepath.Abs(current)
		go system(Shell, absolute)
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
		system(Shell, absolute)
	}
}

func Zero() {
	fmt.Println("-- Phase 0: preliminary system setup.")
	RunEachIn(ZeroDir)
	hostname()
}

func One() {
	fmt.Println("-- Phase 1: launching userspace coroutines.")
	go StartEachIn(OneDirOnce)
	StartLoopEachIn(OneDirRepeat)
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
	Zero()
	One()
	select {}
}
