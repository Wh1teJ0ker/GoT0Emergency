//go:build !windows
// +build !windows

package pdh

import "fmt"

type PDHQuery struct {}

func NewProcessorQueueLengthQuery() (*PDHQuery, error) {
	return nil, fmt.Errorf("not supported on unix")
}

func NewProcessorUtilityQuery() (*PDHQuery, error) {
	return nil, fmt.Errorf("not supported on unix")
}

func NewPDHQuery(counterPath string) (*PDHQuery, error) {
	return nil, fmt.Errorf("not supported on unix")
}

func (q *PDHQuery) Collect() (float64, error) {
	return 0, fmt.Errorf("not supported on unix")
}
