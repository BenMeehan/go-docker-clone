package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	fmt.Println("Logs from your program will appear here!")

	command := os.Args[3]
	args := os.Args[4:]

	cmd := exec.Command(command, args...)

	// Redirect stdout and stderr to parent process
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
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
