//go:build !windows
// +build !windows

// Package pdh provides Windows Performance Data Helper (PDH) API wrapper
// Used for collecting system performance metrics on Windows platforms
// This is a stub implementation for non-Windows platforms
package pdh

import "GoT0Emergency/internal/pkg/log"

// PDHQuery is a stub for non-Windows platforms
type PDHQuery struct{}

// NewProcessorQueueLengthQuery is a stub for non-Windows platforms
func NewProcessorQueueLengthQuery() (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

// NewProcessorUtilityQuery is a stub for non-Windows platforms
func NewProcessorUtilityQuery() (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

// NewPDHQuery is a stub for non-Windows platforms
func NewPDHQuery(counterPath string) (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

// Collect is a stub for non-Windows platforms
func (q *PDHQuery) Collect() (float64, error) {
	return 0, log.Errorf("not supported on unix")
}
