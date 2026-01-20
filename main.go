package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

const lockFile = "/tmp/soglock.tmp"

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}

	err := syscall.Kill(pid, 0)
	return err == nil
}

func readLockPID() int {
	data, err := os.ReadFile(lockFile)
	if err != nil {
		return 0
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0
	}

	return pid
}

func writeLockPID() error {
	return os.WriteFile(lockFile, []byte(strconv.Itoa(os.Getpid())), 0644)
}

func removeLock() {
	_ = os.Remove(lockFile)
}

func setupCleanup() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<- c
		removeLock()
		os.Exit(1)
	}()
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cmd := os.Args[1]

	switch cmd {
	case "sign":
		handleSign(os.Args[2:])
	case "status":
		handleStatus()
	case "unlock":
		handleUnlock()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(2)
	}
}

func handleSign(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: siglock sign [codesign args...]")
		os.Exit(2)
	}

	pid := readLockPID()
	if pid != 0 && isProcessAlive(pid) {
		fmt.Fprintln(os.Stderr, "blocked")
		os.Exit(1)
	}

	removeLock()

	if err := writeLockPID(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to acquire lock: %v\n", err)
		os.Exit(2)
	}

	setupCleanup()

	codesignPath, err := exec.LookPath("codesign")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: 'codesign' not found in PATH")
		removeLock()
		os.Exit(2)
	}

	cmd := exec.Command(codesignPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			exitCode = waitStatus.ExitStatus()
		} else {
			exitCode = 2
		}
	}

	removeLock()
	os.Exit(exitCode)
}

func handleStatus() {
	pid := readLockPID()
	if pid != 0 && isProcessAlive(pid) {
		fmt.Printf("blocked (%d)\n", pid)
		os.Exit(1)
	}
	fmt.Println("free")
	os.Exit(0)
}

func handleUnlock() {
	pid := readLockPID()
	if pid == 0 {
		fmt.Println("was not locked")
	} else {
		fmt.Printf("unlocking (was locked by PID %d)\n", pid)
	}
	removeLock()
	os.Exit(0)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: siglock {sign|status|unlock}
sign   — run codesign with mutual exclusion
status — check if signing is currently blocked
unlock — force-remove lock (use if stuck)
`)
}
