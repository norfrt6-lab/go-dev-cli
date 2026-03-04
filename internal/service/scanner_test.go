package service

import (
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewScanner_DefaultTimeout(t *testing.T) {
	s := NewScanner(0)
	assert.NotNil(t, s)
}

func TestNewScanner_CustomTimeout(t *testing.T) {
	s := NewScanner(500 * time.Millisecond)
	assert.NotNil(t, s)
}

func TestScanner_ScanPort_Open(t *testing.T) {
	// Start a test TCP listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	s := NewScanner(1 * time.Second)

	result := s.ScanPort("127.0.0.1", port)
	assert.True(t, result.Open)
	assert.Equal(t, port, result.Port)
	assert.Greater(t, result.Latency, time.Duration(0))
}

func TestScanner_ScanPort_Closed(t *testing.T) {
	s := NewScanner(200 * time.Millisecond)

	// Use a port that's very unlikely to be open
	result := s.ScanPort("127.0.0.1", 59999)
	assert.False(t, result.Open)
}

func TestScanner_ScanPorts(t *testing.T) {
	// Start two test listeners
	ln1, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln1.Close()

	ln2, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln2.Close()

	port1 := ln1.Addr().(*net.TCPAddr).Port
	port2 := ln2.Addr().(*net.TCPAddr).Port

	s := NewScanner(1 * time.Second)
	results := s.ScanPorts("127.0.0.1", []int{port1, port2, 59998})

	// Should find at least the two open ports
	assert.GreaterOrEqual(t, len(results), 2)

	// Results should be sorted by port
	for i := 1; i < len(results); i++ {
		assert.Less(t, results[i-1].Port, results[i].Port)
	}
}

func TestScanner_ScanPorts_NoneOpen(t *testing.T) {
	s := NewScanner(200 * time.Millisecond)
	results := s.ScanPorts("127.0.0.1", []int{59996, 59997, 59998})
	assert.Empty(t, results)
}

func TestScanner_ScanCommonPorts(t *testing.T) {
	s := NewScanner(100 * time.Millisecond)
	// Just verify it runs without panic — results depend on what's running
	results := s.ScanCommonPorts("127.0.0.1")
	_ = results
}
