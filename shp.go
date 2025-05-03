package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: shp <cmd> [options]")
		return
	}

	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	case "child":
		child(os.Args[2:])
	}
}

func run(args []string) {
	fargs := append([]string{"child"}, args...)
	cmd := exec.Command("/proc/self/exe", fargs...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
	}
	handle(cmd.Run())
}
func child(args []string) {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	newRoot := "/home/ubuntu"
	handle(syscall.Chroot(newRoot))
	handle(syscall.Chdir("/"))
	handle(syscall.Mount("proc", "proc", "proc", 0, ""))

	handle(cmd.Run())
}

func handle(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}
