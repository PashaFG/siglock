package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runSiglock(t *testing.T, lockPath string, args ...string) *exec.Cmd {
	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "TEST_LOCK_PATH="+lockPath)
	return cmd
}

func TestMain(m *testing.M) {
	if lockPath := os.Getenv("TEST_LOCK_PATH"); lockPath != "" {
		lockFilePath = lockPath
		main()
		return
	}

	os.Exit(m.Run())
}

func TestStatus_Free(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "siglock.lock")

	cmd := runSiglock(t, lockPath, "status")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("status failed: %v, output: %s", err, output)
	}

	actual := strings.TrimSpace(string(output))
	if actual != "free" {
		t.Errorf("expected 'free\\n', got %q", string(output))
	}
}

func TestStatus_Blocked(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "siglock.lock")

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := runSiglock(t, lockPath, "status")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected non-zero exit code, got success")
	}
	expected := fmt.Sprintf("blocked (%d)", os.Getpid())
	actual := strings.TrimSpace(string(output))
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, string(output))
	}
}

func TestUnlock(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "siglock.lock")

	if err := os.WriteFile(lockPath, []byte("12345"), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := runSiglock(t, lockPath, "unlock")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("unlock failed: %v, output: %s", err, output)
	}
	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("lock file was not removed")
	}
}

func TestSign_BlockedByLiveProcess(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "siglock.lock")

	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := runSiglock(t, lockPath, "sign", "--help")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Fatalf("expected blocked error, got success")
	}

	actual := strings.TrimSpace(string(output))
	if actual != "blocked" {
		t.Errorf("expected 'blocked\\n', got %q", string(output))
	}
}

func TestSign_IgnoresDeadProcess(t *testing.T) {
	tmpDir := t.TempDir()
	lockPath := filepath.Join(tmpDir, "siglock.lock")

	deadPID := 999999
	if err := os.WriteFile(lockPath, []byte(strconv.Itoa(deadPID)), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := runSiglock(t, lockPath, "sign", "--help")
	output, _ := cmd.CombinedOutput()

	actual := strings.TrimSpace(string(output))
	if actual == "blocked" {
		t.Fatal("should not be blocked by dead PID")
	}

	if _, err := os.Stat(lockPath); !os.IsNotExist(err) {
		t.Error("lock file should be removed after dead PID cleanup")
	}
}
