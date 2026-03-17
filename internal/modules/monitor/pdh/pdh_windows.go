//go:build windows
// +build windows

// Package pdh provides Windows Performance Data Helper (PDH) API wrapper
// Used for collecting system performance metrics on Windows platforms
package pdh

import (
	"GoT0Emergency/internal/pkg/log"
	"syscall"
	"unsafe"
)

var (
	modpdh = syscall.NewLazyDLL("pdh.dll")

	procPdhOpenQuery                = modpdh.NewProc("PdhOpenQueryW")
	procPdhAddEnglishCounter        = modpdh.NewProc("PdhAddEnglishCounterW")
	procPdhCollectQueryData         = modpdh.NewProc("PdhCollectQueryData")
	procPdhGetFormattedCounterValue = modpdh.NewProc("PdhGetFormattedCounterValue")
)

// PDHQuery represents a Windows PDH performance counter query
type PDHQuery struct {
	query   syscall.Handle // Query handle
	counter syscall.Handle // Counter handle
}

// NewPDHQuery creates a new PDH query for the specified counter path
// counterPath: PDH counter path (e.g., "\System\Processor Queue Length")
// Returns: PDHQuery instance and error if counter cannot be opened
func NewPDHQuery(counterPath string) (*PDHQuery, error) {
	var query syscall.Handle
	// 0, 0, &query
	r, _, err := procPdhOpenQuery.Call(0, 0, uintptr(unsafe.Pointer(&query)))
	if r != 0 {
		return nil, log.Errorf("PdhOpenQueryW failed: %v", err)
	}

	var counter syscall.Handle
	path, _ := syscall.UTF16PtrFromString(counterPath)

	// query, path, userData, &counter
	r, _, err = procPdhAddEnglishCounter.Call(
		uintptr(query),
		uintptr(unsafe.Pointer(path)),
		0,
		uintptr(unsafe.Pointer(&counter)),
	)
	if r != 0 {
		return nil, log.Errorf("PdhAddEnglishCounterW failed for %s: %v", counterPath, err)
	}

	// Initial collection
	procPdhCollectQueryData.Call(uintptr(query))

	return &PDHQuery{query: query, counter: counter}, nil
}

// NewProcessorQueueLengthQuery creates a PDH query for processor queue length
// Returns: PDHQuery for "\System\Processor Queue Length" counter
func NewProcessorQueueLengthQuery() (*PDHQuery, error) {
	return NewPDHQuery("\\System\\Processor Queue Length")
}

// NewProcessorUtilityQuery creates a PDH query for processor utility percentage
// Attempts to use "Processor Utility" counter (Task Manager style) first
// Falls back to "Processor Time" on older Windows versions
// Returns: PDHQuery instance and error if both counters fail
func NewProcessorUtilityQuery() (*PDHQuery, error) {
	// Try Processor Information first (Win8+)
	// Note: We use _Total instance
	q, err := NewPDHQuery("\\Processor Information(_Total)\\% Processor Utility")
	if err == nil {
		return q, nil
	}

	// Fallback to traditional Processor Time
	return NewPDHQuery("\\Processor(_Total)\\% Processor Time")
}

// Collect collects the current counter value
// Returns: counter value as float64 and error if collection fails
func (q *PDHQuery) Collect() (float64, error) {
	r, _, err := procPdhCollectQueryData.Call(uintptr(q.query))
	if r != 0 {
		return 0, log.Errorf("PdhCollectQueryData failed: %v", err)
	}

	var value PDH_FMT_COUNTERVALUE_DOUBLE
	var type_ uint32

	// PDH_FMT_DOUBLE = 0x00000200
	r, _, err = procPdhGetFormattedCounterValue.Call(
		uintptr(q.counter),
		uintptr(0x00000200),
		uintptr(unsafe.Pointer(&type_)),
		uintptr(unsafe.Pointer(&value)),
	)
	if r != 0 {
		return 0, log.Errorf("PdhGetFormattedCounterValue failed: %v", err)
	}

	return value.DoubleValue, nil
}

// PDH_FMT_COUNTERVALUE_DOUBLE represents a PDH counter value in double-precision format
// Layout matches Windows PDH_FMT_COUNTERVALUE structure for 64-bit:
// DWORD CStatus; + 4 bytes padding + double doubleValue
type PDH_FMT_COUNTERVALUE_DOUBLE struct {
	CStatus     uint32  // Counter status code
	_           uint32  // Padding for alignment
	DoubleValue float64 // Counter value
}
