package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
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

func mountSystem() {
	syscall.Mount("proc", "/proc", "proc", 0x4|0x2|0x8, "")
	syscall.Mount("sys", "/sys", "sysfs", 0x4|0x2|0x8, "")
	syscall.Mount("run", "/run", "tmpfs", 0x4|0x2, "mode=0755")
	syscall.Mount("dev", "/dev", "devtmpfs", 0x2, "mode=0755")
}

func hostname() {
	host, err := ioutil.ReadFile(HostnameUser)
	if err != nil {
		host = []byte("linux")
	}
	ioutil.WriteFile(HostnameKernel, host, 0644)
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

func preserveEntropy() {
	entropy := make([]byte, 1024)
	rand.Read(entropy)
	ioutil.WriteFile(EntropyReserve, entropy, 0644)
}

func injectEntropy() {
	entropy, err := ioutil.ReadFile(EntropyReserve)
	if err != nil {
		return
	}
	ioutil.WriteFile(EntropyKernel, entropy, 0644)
}

func Zero() {
	fmt.Println("-- Phase 0: preliminary system setup.")
	mountSystem()
	injectEntropy()
	hostname()
	RunEachIn(ZeroDir)
}

func One() {
	fmt.Println("-- Phase 1: launching userspace coroutines.")
	go StartEachIn(OneDirOnce)
	StartLoopEachIn(OneDirRepeat)
}

func Two() {
	fmt.Println("-- Phase 2: reboot triggered, shutdown hooks.")
	preserveEntropy()
	RunEachIn(TwoDir)
}

func main() {
	if os.Getpid() != 1 {
		fmt.Println("Not run as pid 1.")
		os.Exit(1)
	}
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
	for {
		syscall.Wait4(-1, nil, 0, nil)
	}
}
