//go:build windows
// +build windows

package pdh

import (
	"fmt"
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

type PDHQuery struct {
	query   syscall.Handle
	counter syscall.Handle
}

// NewPDHQuery creates a new query for the specified counter path
func NewPDHQuery(counterPath string) (*PDHQuery, error) {
	var query syscall.Handle
	// 0, 0, &query
	r, _, err := procPdhOpenQuery.Call(0, 0, uintptr(unsafe.Pointer(&query)))
	if r != 0 {
		return nil, fmt.Errorf("PdhOpenQueryW failed: %v", err)
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
		return nil, fmt.Errorf("PdhAddEnglishCounterW failed for %s: %v", counterPath, err)
	}

	// Initial collection
	procPdhCollectQueryData.Call(uintptr(query))

	return &PDHQuery{query: query, counter: counter}, nil
}

func NewProcessorQueueLengthQuery() (*PDHQuery, error) {
	return NewPDHQuery("\\System\\Processor Queue Length")
}

// NewProcessorUtilityQuery attempts to open the Processor Utility counter (Task Manager style)
// Fallbacks to Processor Time if Utility is not available (older Windows)
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

func (q *PDHQuery) Collect() (float64, error) {
	r, _, err := procPdhCollectQueryData.Call(uintptr(q.query))
	if r != 0 {
		return 0, fmt.Errorf("PdhCollectQueryData failed: %v", err)
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
		return 0, fmt.Errorf("PdhGetFormattedCounterValue failed: %v", err)
	}

	return value.DoubleValue, nil
}

// PDH_FMT_COUNTERVALUE structure layout:
// DWORD CStatus;
// union { ... double doubleValue ... };
// On 64-bit Go, struct fields are aligned.
// uint32 is 4 bytes. float64 is 8 bytes.
// So there will be 4 bytes padding after CStatus.
type PDH_FMT_COUNTERVALUE_DOUBLE struct {
	CStatus     uint32
	_           uint32 // Padding
	DoubleValue float64
}
