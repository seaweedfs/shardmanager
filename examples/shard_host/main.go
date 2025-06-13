package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/seaweedfs/shardmanager/policy"
)

// ShardHost represents a node that can host shards
type ShardHost struct {
	ID       string
	Shards   map[string]*Shard
	mu       sync.RWMutex
	metrics  map[string]float64
	stopChan chan struct{}
}

// Shard represents a shard hosted on this node
type Shard struct {
	ID       string
	Data     []byte
	Metadata map[string]interface{}
}

// NewShardHost creates a new shard host
func NewShardHost(id string) *ShardHost {
	return &ShardHost{
		ID:       id,
		Shards:   make(map[string]*Shard),
		metrics:  make(map[string]float64),
		stopChan: make(chan struct{}),
	}
}

// StartMetricsCollection starts collecting metrics for this host
func (h *ShardHost) StartMetricsCollection() {
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				h.updateMetrics()
			case <-h.stopChan:
				return
			}
		}
	}()
}

// updateMetrics updates the host's metrics
func (h *ShardHost) updateMetrics() {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Simulate CPU usage based on number of shards
	cpuUsage := 20.0 + float64(len(h.Shards))*10.0
	h.metrics["cpu_usage"] = cpuUsage

	// Simulate memory usage based on total shard size
	memoryUsage := 30.0
	for _, shard := range h.Shards {
		memoryUsage += float64(len(shard.Data)) / 1024.0 // Convert to KB
	}
	h.metrics["memory_usage"] = memoryUsage
}

// GetMetrics returns the current metrics
func (h *ShardHost) GetMetrics() map[string]float64 {
	h.mu.RLock()
	defer h.mu.RUnlock()

	metrics := make(map[string]float64)
	for k, v := range h.metrics {
		metrics[k] = v
	}
	return metrics
}

// AddShard adds a new shard to this host
func (h *ShardHost) AddShard(shardID string, data []byte, metadata map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Shards[shardID]; exists {
		return fmt.Errorf("shard %s already exists", shardID)
	}

	h.Shards[shardID] = &Shard{
		ID:       shardID,
		Data:     data,
		Metadata: metadata,
	}

	log.Printf("Added shard %s to host %s", shardID, h.ID)
	return nil
}

// RemoveShard removes a shard from this host
func (h *ShardHost) RemoveShard(shardID string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.Shards[shardID]; !exists {
		return fmt.Errorf("shard %s not found", shardID)
	}

	delete(h.Shards, shardID)
	log.Printf("Removed shard %s from host %s", shardID, h.ID)
	return nil
}

// GetShard returns a shard's data and metadata
func (h *ShardHost) GetShard(shardID string) (*Shard, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	shard, exists := h.Shards[shardID]
	if !exists {
		return nil, fmt.Errorf("shard %s not found", shardID)
	}

	return shard, nil
}

// HandleMigrateShard handles a shard migration action
func (h *ShardHost) HandleMigrateShard(action policy.Action) error {
	source, ok := action.Constraints["source"].(string)
	if !ok {
		return fmt.Errorf("invalid source in migration action")
	}

	target, ok := action.Constraints["target"].(string)
	if !ok {
		return fmt.Errorf("invalid target in migration action")
	}

	// If this is the source host, prepare the shard for migration
	if source == h.ID {
		// In a real implementation, this would:
		// 1. Stop accepting writes to the shard
		// 2. Take a snapshot of the shard
		// 3. Send the snapshot to the target host
		// 4. Remove the shard from this host
		log.Printf("Preparing shard for migration from %s to %s", source, target)
		return nil
	}

	// If this is the target host, receive the shard
	if target == h.ID {
		// In a real implementation, this would:
		// 1. Receive the shard data from the source
		// 2. Start accepting writes to the shard
		log.Printf("Receiving shard from %s to %s", source, target)
		return nil
	}

	return nil
}

func main() {
	// Create a new shard host
	host := NewShardHost("node1")
	host.StartMetricsCollection()

	// Add some test shards
	for i := 0; i < 3; i++ {
		shardID := fmt.Sprintf("shard-%d", i)
		data := make([]byte, 1024) // 1KB of test data
		metadata := map[string]interface{}{
			"created_at": time.Now(),
			"size":       len(data),
		}
		if err := host.AddShard(shardID, data, metadata); err != nil {
			log.Fatalf("Failed to add shard: %v", err)
		}
	}

	// Start an HTTP server to expose metrics and handle actions
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := host.GetMetrics()
		fmt.Fprintf(w, "Host %s metrics:\n", host.ID)
		for k, v := range metrics {
			fmt.Fprintf(w, "%s: %.2f\n", k, v)
		}
	})

	http.HandleFunc("/shards", func(w http.ResponseWriter, r *http.Request) {
		host.mu.RLock()
		defer host.mu.RUnlock()

		fmt.Fprintf(w, "Host %s shards:\n", host.ID)
		for id, shard := range host.Shards {
			fmt.Fprintf(w, "Shard %s: %d bytes\n", id, len(shard.Data))
		}
	})

	// Start the server
	go func() {
		log.Printf("Starting server on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Print metrics periodically
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := host.GetMetrics()
				log.Printf("Current metrics: CPU=%.1f%%, Memory=%.1f%%",
					metrics["cpu_usage"],
					metrics["memory_usage"])
			case <-host.stopChan:
				return
			}
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	close(host.stopChan)
	log.Println("Shard host stopped")
}
