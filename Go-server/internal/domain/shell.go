package domain

import (
	"fmt"

	"os"
	"os/exec"
	"time"

	"github.com/creack/pty"
)

type Shell struct {
	pty  *os.File
	cmd  *exec.Cmd
	rows uint16
	cols uint16
}

// Spawn a new bash shell with PTY and returns a Shell instance
func NewShell(rows, cols uint16) (*Shell, error) {
	cmd := exec.Command("/bin/bash")

	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	winsize := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}
	if err := pty.Setsize(ptyFile, winsize); err != nil {
		ptyFile.Close()
		cmd.Process.Kill()
		return nil, fmt.Errorf("failed to set PTY size: %w", err)
	}

	return &Shell{
		pty:  ptyFile,
		cmd:  cmd,
		rows: rows,
		cols: cols,
	}, nil
}

// Reads output from the PTY (shell output)
func (s *Shell) Read(p []byte) (int, error) {
	return s.pty.Read(p)
}

// Writes input to the PTY (shell stdin)
func (s *Shell) Write(p []byte) (int, error) {
	return s.pty.Write(p)
}

// Resize updates the terminal dimensions
func (s *Shell) Resize(rows, cols uint16) error {
	winsize := &pty.Winsize{
		Rows: rows,
		Cols: cols,
	}

	if err := pty.Setsize(s.pty, winsize); err != nil {
		return fmt.Errorf("failed to resize PTY: %w", err)
	}

	s.rows = rows
	s.cols = cols

	return nil
}

// Close gracefully terminates the shell session
func (s *Shell) Close() error {
	var firstErr error

	if err := s.pty.Close(); err != nil {
		firstErr = fmt.Errorf("failed to close PTY: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- s.cmd.Wait()
	}()

	select {
	case err := <-done:
		// Process exited normally
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("process exited with error: %w", err)
		}
	case <-time.After(5 * time.Second):
		// Timeout: force kill the process
		if err := s.cmd.Process.Kill(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to kill process: %w", err)
		}
	}

	return firstErr
}
