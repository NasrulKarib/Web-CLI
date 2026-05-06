package usecase

import (
	"context"
	"errors"
	"sync"

	"websocket-server/internal/domain"
)

// ShellService manages PTY lifecycle at business logic level.
type ShellService struct {
	shell *domain.Shell // Current active shell (nil if no session).
	mu    sync.Mutex    // Protects shell pointer mutations (Start, Stop, Resize).
}

// NewShellService creates and returns a new ShellService instance.
func NewShellService() *ShellService {
	return &ShellService{
		shell: nil,
	}
}

// Start creates a new shell session with the specified terminal dimensions.
func (s *ShellService) Start(_ context.Context, rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell != nil {
		return errors.New("shell session already active")
	}

	shell, err := domain.NewShell(rows, cols)
	if err != nil {
		return err
	}

	s.shell = shell
	return nil
}

// Write sends data to the active shell session.
// No mutex: PTY fd supports concurrent read/write. Write only accesses
// s.shell which is stable during the handler's lifetime (Stop runs after goroutines exit).
func (s *ShellService) Write(data []byte) (int, error) {
	if s.shell == nil {
		return 0, errors.New("no active shell session")
	}

	return s.shell.Write(data)
}

// Read retrieves output from the active shell session (blocking).
// Prefer ReadContext for goroutine-safe cancellation.
func (s *ShellService) Read() ([]byte, error) {
	if s.shell == nil {
		return nil, errors.New("no active shell session")
	}

	buffer := make([]byte, 4096)
	n, err := s.shell.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:n], nil
}

// ReadContext reads PTY output with context-aware cancellation.
func (s *ShellService) ReadContext(ctx context.Context) ([]byte, error) {
	if s.shell == nil {
		return nil, errors.New("no active shell session")
	}

	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)

	go func() {
		buf := make([]byte, 4096)
		n, err := s.shell.Read(buf)
		if err != nil {
			ch <- result{nil, err}
			return
		}
		ch <- result{buf[:n], nil}
	}()

	select {
	case r := <-ch:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Resize updates the terminal dimensions of the active shell.
func (s *ShellService) Resize(rows, cols uint16) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == nil {
		return errors.New("no active shell session")
	}

	return s.shell.Resize(rows, cols)
}

// Stop terminates the active shell session (idempotent).
func (s *ShellService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.shell == nil {
		return nil
	}

	err := s.shell.Close()
	s.shell = nil
	return err
}

// IsActive returns true if a shell session is currently active.
func (s *ShellService) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.shell != nil
}
