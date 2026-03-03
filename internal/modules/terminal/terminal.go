package terminal

type Terminal interface {
	// Read() should be handled internally by the implementation to stream data to the UI
	Write(data []byte) (int, error)
	Resize(rows, cols int) error
	Close() error
	Wait() error // Wait for process to exit
}
