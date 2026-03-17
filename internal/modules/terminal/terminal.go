// Package terminal provides terminal interface definitions
package terminal

// Terminal interface for terminal implementations
// Each terminal session (local or SSH) implements this interface
type Terminal interface {
	// Write sends data to the terminal (user input)
	Write(data []byte) (int, error)
	// Resize changes the terminal dimensions
	Resize(rows, cols int) error
	// Close closes the terminal session
	Close() error
	// Wait waits for the terminal process to exit
	Wait() error
}
