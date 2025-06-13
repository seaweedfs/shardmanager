package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// HostInfo represents a shard host
type HostInfo struct {
	ID   string
	Addr string // e.g. http://localhost:8080
}

// MigrationRequest represents a migration action
type MigrationRequest struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Shard  string `json:"shard"`
}

func fetchMetrics(host HostInfo) (map[string]float64, error) {
	resp, err := http.Get(host.Addr + "/metrics")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	metrics := make(map[string]float64)
	for _, line := range bytes.Split(body, []byte("\n")) {
		lineStr := strings.TrimSpace(string(line))
		if lineStr == "" {
			continue
		}
		parts := strings.SplitN(lineStr, ":", 2)
		if len(parts) != 2 {
			continue
		}
		k := strings.TrimSpace(parts[0])
		vStr := strings.TrimSpace(parts[1])
		var v float64
		_, err := fmt.Sscanf(vStr, "%f", &v)
		if err != nil {
			continue
		}
		metrics[k] = v
	}
	return metrics, nil
}

func fetchShards(host HostInfo) ([]string, error) {
	resp, err := http.Get(host.Addr + "/shards")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var shards []string
	for _, line := range bytes.Split(body, []byte("\n")) {
		lineStr := strings.TrimSpace(string(line))
		if lineStr == "" {
			continue
		}
		parts := strings.SplitN(lineStr, ":", 2)
		if len(parts) != 2 {
			continue
		}
		shardID := strings.TrimSpace(strings.TrimPrefix(parts[0], "Shard "))
		shards = append(shards, shardID)
	}
	return shards, nil
}

func triggerMigration(source, target HostInfo, shardID string) error {
	migration := MigrationRequest{
		Source: source.ID,
		Target: target.ID,
		Shard:  shardID,
	}
	// In a real system, this would be a POST to a migration endpoint
	// For this example, just log the action
	logPrintf("Triggering migration: %+v\n", migration)
	return nil
}

// logPrintf is used to allow test interception of log output
var logPrintf = fmt.Printf

func main() {
	// Start the metrics server for node1
	go func() {
		http.HandleFunc("/node1/metrics", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "cpu_usage: 50.0\nmemory_usage: 33.0\n")
		})
		http.HandleFunc("/node1/shards", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Shard 1: active\nShard 2: active\n")
		})
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	// Start the metrics server for node2
	go func() {
		http.HandleFunc("/node2/metrics", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "cpu_usage: 20.0\nmemory_usage: 25.0\n")
		})
		http.HandleFunc("/node2/shards", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "Shard 3: active\n")
		})
		log.Fatal(http.ListenAndServe(":8081", nil))
	}()

	// Wait for the servers to start
	time.Sleep(1 * time.Second)

	hosts := []HostInfo{
		{ID: "node1", Addr: "http://localhost:8080"},
		{ID: "node2", Addr: "http://localhost:8081"},
	}

	// Wait for hosts to be up
	log.Println("Waiting for hosts to be ready...")
	time.Sleep(2 * time.Second)

	// Fetch metrics and shards from each host
	for _, host := range hosts {
		metrics, err := fetchMetrics(host)
		if err != nil {
			log.Printf("Failed to fetch metrics from %s: %v", host.ID, err)
			continue
		}
		log.Printf("Host %s metrics: %+v", host.ID, metrics)

		shards, err := fetchShards(host)
		if err != nil {
			log.Printf("Failed to fetch shards from %s: %v", host.ID, err)
			continue
		}
		log.Printf("Host %s shards: %v", host.ID, shards)
	}

	// Example orchestration: if node1 has high CPU, migrate a shard to node2
	source := hosts[0]
	target := hosts[1]
	metrics, err := fetchMetrics(source)
	if err != nil {
		log.Fatalf("Failed to fetch metrics from %s: %v", source.ID, err)
	}
	if metrics["cpu_usage"] > 30.0 {
		shards, err := fetchShards(source)
		if err != nil || len(shards) == 0 {
			log.Printf("No shards to migrate from %s", source.ID)
			os.Exit(0)
		}
		shardID := shards[0]
		if err := triggerMigration(source, target, shardID); err != nil {
			log.Printf("Migration failed: %v", err)
		} else {
			log.Printf("Migration triggered for shard %s from %s to %s", shardID, source.ID, target.ID)
		}
	} else {
		log.Printf("No migration needed. CPU usage on %s is %.2f", source.ID, metrics["cpu_usage"])
	}

	// Simulate some orchestration logic
	time.Sleep(10 * time.Second)
	fmt.Println("Orchestration example completed.")
}
