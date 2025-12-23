package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	oldRootDir = ".old_root"
	procFS     = "proc"
)

// Isolator defines filesystem isolation strategies
type Isolator interface {
	Isolate(rootfs string) error
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: shp run <rootfs_path> <cmd> [options]")
		return
	}

	switch os.Args[1] {
	case "run":
		run(os.Args[2:])
	case "child":
		child(os.Args[2:])
	default:
		fmt.Println("usage: shp run <rootfs_path> <cmd> [options]")
	}
}

func run(args []string) {
	if len(args) < 2 {
		fmt.Println("usage: shp run <rootfs_path> <cmd> [options]")
		os.Exit(1)
	}

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
	if len(args) < 2 {
		fmt.Println("usage: shp child <rootfs_path> <cmd> [options]")
		os.Exit(1)
	}

	rootfs := args[0]
	cmdArgs := args[1:]

	handle(validateRootfs(rootfs))
	binPath := getCmdPath(cmdArgs[0])

	cmd := exec.Command(binPath, cmdArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Try pivot_root first, fall back to chroot
	err := (&PivotRootIsolator{}).Isolate(rootfs)
	if err != nil {
		fmt.Printf("pivot_root failed: %v\nFalling back to chroot...\n", err)
		handle((&ChrootIsolator{}).Isolate(rootfs))
	}

	handle(mountProc())
	handle(cmd.Run())
}

// PivotRootIsolator uses pivot_root for filesystem isolation
type PivotRootIsolator struct{}

func (p *PivotRootIsolator) Isolate(rootfs string) error {
	absNewRoot, err := filepath.Abs(rootfs)
	if err != nil {
		return fmt.Errorf("cannot get absolute path for %s: %w", rootfs, err)
	}

	oldRoot := filepath.Join(absNewRoot, oldRootDir)
	if err := os.MkdirAll(oldRoot, 0700); err != nil {
		return fmt.Errorf("cannot create old_root directory: %w", err)
	}

	if err := syscall.PivotRoot(absNewRoot, oldRoot); err != nil {
		return fmt.Errorf("pivot_root failed: %w", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir to / failed after pivot_root: %w", err)
	}

	// Unmount old root - non-critical, log but don't fail
	if err := syscall.Unmount("/"+oldRootDir, syscall.MNT_DETACH); err != nil {
		fmt.Printf("Warning: unmounting old root failed: %v\n", err)
	}

	// Remove old root directory - non-critical, log but don't fail
	if err := os.Remove("/" + oldRootDir); err != nil {
		fmt.Printf("Warning: removing old root directory failed: %v\n", err)
	}

	fmt.Println("Successfully using pivot_root")
	return nil
}

// ChrootIsolator uses chroot for filesystem isolation (fallback)
type ChrootIsolator struct{}

func (c *ChrootIsolator) Isolate(rootfs string) error {
	if err := syscall.Chroot(rootfs); err != nil {
		return fmt.Errorf("chroot failed: %w", err)
	}
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir to / failed after chroot: %w", err)
	}
	fmt.Println("Using chroot for filesystem isolation")
	return nil
}

func validateRootfs(rootfs string) error {
	if _, err := os.Stat(rootfs); err != nil {
		return fmt.Errorf("rootfs path does not exist: %s: %w", rootfs, err)
	}
	return nil
}
func getCmdPath(cmdPath string) string {
	splits := strings.Split(cmdPath, "/")
	if len(splits) > 1 {
		fmt.Printf("WARNING! Absolute path resolution for [%s] will be done based on the new rootfs (inside container).\n", cmdPath)
		return cmdPath
	}
	fmt.Printf("INFO: Resolving command [%s] inside /bin of the new rootfs.\n", cmdPath)
	return filepath.Join("/bin/", cmdPath)
}

func mountProc() error {
	return syscall.Mount(procFS, procFS, procFS, 0, "")
}

func handle(err error) {
	if err != nil {
		fmt.Printf("\n%s\n", err.Error())
		os.Exit(1)
	}
}
