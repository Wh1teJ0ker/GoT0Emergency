//go:build !windows
// +build !windows

package pdh

import "GoT0Emergency/internal/pkg/log"

type PDHQuery struct{}

func NewProcessorQueueLengthQuery() (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

func NewProcessorUtilityQuery() (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

func NewPDHQuery(counterPath string) (*PDHQuery, error) {
	return nil, log.Errorf("not supported on unix")
}

func (q *PDHQuery) Collect() (float64, error) {
	return 0, log.Errorf("not supported on unix")
}
