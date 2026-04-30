package usecase

import (
	"context"
	"errors"
	"sync"

	"websocket-server/internal/domain"
)

// ShellService manages PTY lifecycle at business logic level
type ShellService struct {
	shell *domain.Shell // Current active shell (nil if no session)
	mu    sync.Mutex    // Protect concurrent access
}

// NewShellService creates and returns a new ShellService instance
func NewShellService() *ShellService {
	return &ShellService{
		shell: nil,
	}
}

// Start creates a new shell session with the specified terminal dimensions
func (s *ShellService) Start(ctx context.Context, rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if session already active
	if s.shell != nil {
		return errors.New("shell session already active")
	}

	// Create new shell
	shell, err := domain.NewShell(rows, cols)
	if err != nil {
		return err
	}

	s.shell = shell
	return nil
}

// Write sends data to the active shell session
func (s *ShellService) Write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == nil {
		return 0, errors.New("no active shell session")
	}

	return s.shell.Write(data)
}

// Read retrieves output from the active shell session
func (s *ShellService) Read() ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == nil {
		return nil, errors.New("no active shell session")
	}

	// Create read buffer (4096 bytes)
	buffer := make([]byte, 4096)
	n, err := s.shell.Read(buffer)
	if err != nil {
		return nil, err
	}

	// Return slice of actual bytes read
	return buffer[:n], nil
}

// Resize updates the terminal dimensions of the active shell
func (s *ShellService) Resize(rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == nil {
		return errors.New("no active shell session")
	}

	return s.shell.Resize(rows, cols)
}

// Stop terminates the active shell session (idempotent)
func (s *ShellService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Already stopped - idempotent behavior
	if s.shell == nil {
		return nil
	}

	err := s.shell.Close()
	s.shell = nil
	return err
}

// IsActive returns true if a shell session is currently active
func (s *ShellService) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.shell != nil
}
