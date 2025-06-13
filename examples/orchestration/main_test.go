package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestOrchestration(t *testing.T) {
	// Simulate two shard hosts with different metrics and shards
	host1Metrics := "cpu_usage: 40.0\nmemory_usage: 50.0\n"
	host1Shards := "Shard shard-1: 1024 bytes\nShard shard-2: 1024 bytes\n"
	host2Metrics := "cpu_usage: 10.0\nmemory_usage: 20.0\n"
	host2Shards := "Shard shard-3: 1024 bytes\n"

	host1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/metrics") {
			fmt.Fprint(w, host1Metrics)
		} else if strings.HasPrefix(r.URL.Path, "/shards") {
			fmt.Fprint(w, host1Shards)
		}
	}))
	defer host1.Close()

	host2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/metrics") {
			fmt.Fprint(w, host2Metrics)
		} else if strings.HasPrefix(r.URL.Path, "/shards") {
			fmt.Fprint(w, host2Shards)
		}
	}))
	defer host2.Close()

	hosts := []HostInfo{
		{ID: "node1", Addr: host1.URL},
		{ID: "node2", Addr: host2.URL},
	}

	// Test fetchMetrics and fetchShards
	metrics1, err := fetchMetrics(hosts[0])
	if err != nil {
		t.Fatalf("fetchMetrics failed: %v", err)
	}
	if metrics1["cpu_usage"] != 40.0 {
		t.Errorf("expected cpu_usage 40.0, got %v", metrics1["cpu_usage"])
	}

	shards1, err := fetchShards(hosts[0])
	if err != nil {
		t.Fatalf("fetchShards failed: %v", err)
	}
	if len(shards1) != 2 {
		t.Errorf("expected 2 shards, got %d", len(shards1))
	}

	metrics2, err := fetchMetrics(hosts[1])
	if err != nil {
		t.Fatalf("fetchMetrics failed: %v", err)
	}
	if metrics2["cpu_usage"] != 10.0 {
		t.Errorf("expected cpu_usage 10.0, got %v", metrics2["cpu_usage"])
	}

	shards2, err := fetchShards(hosts[1])
	if err != nil {
		t.Fatalf("fetchShards failed: %v", err)
	}
	if len(shards2) != 1 {
		t.Errorf("expected 1 shard, got %d", len(shards2))
	}

	// Test orchestration logic: migration should be triggered if node1's cpu_usage > 30
	var migrationTriggered int32
	logPrintf = func(format string, v ...interface{}) (int, error) {
		msg := fmt.Sprintf(format, v...)
		if strings.Contains(msg, "Triggering migration") {
			atomic.StoreInt32(&migrationTriggered, 1)
		}
		return len(msg), nil
	}
	defer func() { logPrintf = fmt.Printf }()

	if metrics1["cpu_usage"] > 30.0 {
		shardID := shards1[0]
		_ = triggerMigration(hosts[0], hosts[1], shardID)
	}

	if atomic.LoadInt32(&migrationTriggered) == 0 {
		t.Errorf("expected migration to be triggered, but it was not")
	}
}
