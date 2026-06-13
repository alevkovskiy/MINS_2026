package client

import (
	"sync/atomic"
)

type GRPCMetrics struct {
	totalCalls       atomic.Int64
	successCalls     atomic.Int64
	failedCalls      atomic.Int64
	timeoutCalls     atomic.Int64
	unavailableCalls atomic.Int64
	invalidArgsCalls atomic.Int64
	internalErrCalls atomic.Int64
}

func NewGRPCMetrics() *GRPCMetrics {
	return &GRPCMetrics{}
}

func (m *GRPCMetrics) RecordCall(success bool, errorCode int) {
	m.totalCalls.Add(1)

	if success {
		m.successCalls.Add(1)
		return
	}

	m.failedCalls.Add(1)

	switch errorCode {
	case 4:
		m.timeoutCalls.Add(1)
	case 14:
		m.unavailableCalls.Add(1)
	case 3:
		m.invalidArgsCalls.Add(1)
	case 13:
		m.internalErrCalls.Add(1)
	}
}

func (m *GRPCMetrics) GetStats() map[string]int64 {
	return map[string]int64{
		"total_calls":       m.totalCalls.Load(),
		"success_calls":     m.successCalls.Load(),
		"failed_calls":      m.failedCalls.Load(),
		"timeout_calls":     m.timeoutCalls.Load(),
		"unavailable_calls": m.unavailableCalls.Load(),
		"invalid_args":      m.invalidArgsCalls.Load(),
		"internal_errors":   m.internalErrCalls.Load(),
	}
}

func (m *GRPCMetrics) GetSuccessRate() float64 {
	total := m.totalCalls.Load()
	if total == 0 {
		return 100.0
	}
	return float64(m.successCalls.Load()) / float64(total) * 100
}
