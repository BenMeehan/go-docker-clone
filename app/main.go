package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// works only on linux
// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	fmt.Println("Logs from your program will appear here!")

	command := os.Args[3]
	args := os.Args[4:]

	// Create a temporary directory for chroot
	chrootDir, err := ioutil.TempDir("", "docker-chroot")
	if err != nil {
		fmt.Printf("Error creating chroot directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(chrootDir) // Clean up the temporary directory

	// Copy the command binary into the chroot directory
	cmdBinary := filepath.Join(chrootDir, filepath.Base(command))
	if err := copyFile(command, cmdBinary); err != nil {
		fmt.Printf("Error copying command binary: %v\n", err)
		os.Exit(1)
	}

	// Chroot into the temporary directory
	if err := syscall.Chroot(chrootDir); err != nil {
		fmt.Printf("Error chrooting into directory: %v\n", err)
		os.Exit(1)
	}

	// Mount /proc to hide other processes and set the process ID to 1
	if err := syscall.Mount("proc", "proc", "proc", 0, ""); err != nil {
		fmt.Printf("Error mounting /proc: %v\n", err)
		os.Exit(1)
	}
	defer syscall.Unmount("/proc", 0) // Unmount /proc when done

	// Redirect stdout and stderr to parent process
	cmd := exec.Command(filepath.Base(command), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set the process ID to 1
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Run the command
	err = cmd.Run()
	if err != nil {
		// Check if the error is an exit error
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Get the exit code from the exit error
			exitCode := exitErr.Sys().(syscall.WaitStatus).ExitStatus()
			// Exit with the same exit code
			os.Exit(exitCode)
		}
		// If it's not an exit error, exit with code 1
		os.Exit(1)
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return nil
}