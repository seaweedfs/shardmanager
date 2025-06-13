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
	port     int
}

// Shard represents a shard hosted on this node
type Shard struct {
	ID       string
	Data     []byte
	Metadata map[string]interface{}
}

// NewShardHost creates a new shard host
func NewShardHost(id string, port int) *ShardHost {
	return &ShardHost{
		ID:       id,
		Shards:   make(map[string]*Shard),
		metrics:  make(map[string]float64),
		stopChan: make(chan struct{}),
		port:     port,
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

	shardID, ok := action.Constraints["shard"].(string)
	if !ok {
		return fmt.Errorf("invalid shard in migration action")
	}

	// If this is the source host, prepare the shard for migration
	if source == h.ID {
		_, err := h.GetShard(shardID)
		if err != nil {
			return err
		}

		// In a real implementation, this would send the shard data to the target
		log.Printf("Migrating shard %s from %s to %s", shardID, source, target)

		// Remove the shard from this host
		return h.RemoveShard(shardID)
	}

	// If this is the target host, receive the shard
	if target == h.ID {
		// In a real implementation, this would receive the shard data from the source
		log.Printf("Receiving shard %s from %s to %s", shardID, source, target)

		// Add the shard to this host (with mock data for this example)
		data := make([]byte, 1024)
		metadata := map[string]interface{}{
			"migrated_at": time.Now(),
			"size":        len(data),
		}
		return h.AddShard(shardID, data, metadata)
	}

	return nil
}

// StartServer starts the HTTP server for this host
func (h *ShardHost) StartServer() {
	http.HandleFunc(fmt.Sprintf("/%s/metrics", h.ID), func(w http.ResponseWriter, r *http.Request) {
		metrics := h.GetMetrics()
		fmt.Fprintf(w, "Host %s metrics:\n", h.ID)
		for k, v := range metrics {
			fmt.Fprintf(w, "%s: %.2f\n", k, v)
		}
	})

	http.HandleFunc(fmt.Sprintf("/%s/shards", h.ID), func(w http.ResponseWriter, r *http.Request) {
		h.mu.RLock()
		defer h.mu.RUnlock()

		fmt.Fprintf(w, "Host %s shards:\n", h.ID)
		for id, shard := range h.Shards {
			fmt.Fprintf(w, "Shard %s: %d bytes\n", id, len(shard.Data))
		}
	})

	go func() {
		addr := fmt.Sprintf(":%d", h.port)
		log.Printf("Starting server for %s on %s", h.ID, addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Failed to start server for %s: %v", h.ID, err)
		}
	}()
}

func main() {
	// Create the first shard host
	host1 := NewShardHost("node1", 8080)
	host1.StartMetricsCollection()
	host1.StartServer()

	// Add some test shards to the first host
	for i := 0; i < 3; i++ {
		shardID := fmt.Sprintf("shard-%d", i)
		data := make([]byte, 1024) // 1KB of test data
		metadata := map[string]interface{}{
			"created_at": time.Now(),
			"size":       len(data),
		}
		if err := host1.AddShard(shardID, data, metadata); err != nil {
			log.Fatalf("Failed to add shard: %v", err)
		}
	}

	// Print metrics periodically for host1
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := host1.GetMetrics()
				log.Printf("Host1 metrics: CPU=%.1f%%, Memory=%.1f%%",
					metrics["cpu_usage"],
					metrics["memory_usage"])
			case <-host1.stopChan:
				return
			}
		}
	}()

	// Start the second host after a delay
	time.Sleep(10 * time.Second)
	log.Println("Starting second host...")
	host2 := NewShardHost("node2", 8081)
	host2.StartMetricsCollection()
	host2.StartServer()

	// Print metrics periodically for host2
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := host2.GetMetrics()
				log.Printf("Host2 metrics: CPU=%.1f%%, Memory=%.1f%%",
					metrics["cpu_usage"],
					metrics["memory_usage"])
			case <-host2.stopChan:
				return
			}
		}
	}()

	// Simulate load balancing after a delay
	time.Sleep(5 * time.Second)
	log.Println("Starting load balancing...")

	// Migrate one shard from host1 to host2
	action := policy.Action{
		Type: "migrate_shard",
		Constraints: map[string]interface{}{
			"source": "node1",
			"target": "node2",
			"shard":  "shard-0",
		},
	}

	if err := host1.HandleMigrateShard(action); err != nil {
		log.Printf("Failed to migrate shard from host1: %v", err)
	}
	if err := host2.HandleMigrateShard(action); err != nil {
		log.Printf("Failed to migrate shard to host2: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Cleanup
	close(host1.stopChan)
	close(host2.stopChan)
	log.Println("Shard hosts stopped")
}
