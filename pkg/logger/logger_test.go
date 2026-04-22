package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestInitAndLogFunctions(t *testing.T) {
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stdout error = %v", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() stderr error = %v", err)
	}

	originalStdout := os.Stdout
	originalStderr := os.Stderr
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	t.Cleanup(func() {
		os.Stdout = originalStdout
		os.Stderr = originalStderr
	})

	instance = nil
	once = sync.Once{}
	Init()

	Info("hello", "key", "value")
	Warn("watch-out")
	Error("boom")
	Fatal("fatal-message")

	_ = stdoutWriter.Close()
	_ = stderrWriter.Close()

	stdout := readAll(t, stdoutReader)
	stderr := readAll(t, stderrReader)

	if !strings.Contains(stdout, "[INFO]") || !strings.Contains(stdout, "hello key=value") {
		t.Fatalf("stdout = %q, want info log content", stdout)
	}
	if !strings.Contains(stdout, "[WARN]") || !strings.Contains(stdout, "watch-out") {
		t.Fatalf("stdout = %q, want warn log content", stdout)
	}
	if !strings.Contains(stderr, "[ERROR]") || !strings.Contains(stderr, "boom") || !strings.Contains(stderr, "fatal-message") {
		t.Fatalf("stderr = %q, want error log content", stderr)
	}
}

func readAll(t *testing.T, reader *os.File) string {
	t.Helper()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("io.Copy() error = %v", err)
	}
	return buf.String()
}
