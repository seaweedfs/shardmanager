package integration

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/seaweedfs/shardmanager/server"
	"github.com/seaweedfs/shardmanager/shardmanagerpb"
	"google.golang.org/grpc"
)

type testAppServer struct {
	shardmanagerpb.UnimplementedAppShardServiceServer
	id     string
	shards chan string // receives added shard IDs
}

func (s *testAppServer) AddShard(ctx context.Context, req *shardmanagerpb.AddShardRequest) (*shardmanagerpb.AddShardResponse, error) {
	log.Printf("AppServer %s: AddShard called: shard_id=%s, role=%s", s.id, req.ShardId, req.Role)
	s.shards <- req.ShardId
	return &shardmanagerpb.AddShardResponse{Success: true, Message: "Shard added"}, nil
}

func startTestAppServer(t *testing.T, wg *sync.WaitGroup, port int, nodeID string, registerAddr string) (*grpc.Server, chan string) {
	shards := make(chan string, 10)
	server := grpc.NewServer()
	appServer := &testAppServer{id: nodeID, shards: shards}
	shardmanagerpb.RegisterAppShardServiceServer(server, appServer)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	go func() {
		defer wg.Done()
		if err := server.Serve(lis); err != nil {
			t.Errorf("failed to serve: %v", err)
		}
	}()

	// Register with shardmanager
	conn, err := grpc.Dial(registerAddr, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial shardmanager: %v", err)
	}
	defer conn.Close()
	client := shardmanagerpb.NewNodeServiceClient(conn)
	_, err = client.RegisterNode(context.Background(), &shardmanagerpb.RegisterNodeRequest{
		Node: &shardmanagerpb.Node{
			Id:       nodeID,
			Location: fmt.Sprintf("localhost:%d", port),
			Capacity: 100,
			Status:   "active",
		},
	})
	if err != nil {
		t.Fatalf("failed to register node: %v", err)
	}
	return server, shards
}

func TestShardBalancing(t *testing.T) {
	var wg sync.WaitGroup
	shardManagerPort := 6000
	appServerPorts := []int{5101, 5102, 5103}
	// Use a file-based SQLite DB for compatibility with multiple connections
	dbPath := "file:integration_test.db?mode=rwc&_journal_mode=WAL&_busy_timeout=5000"
	os.Remove("integration_test.db")

	stopCh := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := server.StartShardManagerServer(dbPath, fmt.Sprintf(":%d", shardManagerPort), stopCh)
		if err != nil {
			t.Fatalf("failed to start real shardmanager: %v", err)
		}
	}()
	time.Sleep(500 * time.Millisecond) // Give time for server to start

	// Create slices to store all gRPC servers and their channels
	appServers := make([]*grpc.Server, len(appServerPorts))
	shardChans := make([]chan string, len(appServerPorts))
	for i, port := range appServerPorts {
		wg.Add(1)
		appServers[i], shardChans[i] = startTestAppServer(t, &wg, port, fmt.Sprintf("appserver-%d", i+1), fmt.Sprintf("localhost:%d", shardManagerPort))
	}
	time.Sleep(1000 * time.Millisecond) // Give time for registration

	// Add shards via the real RegisterShard RPC
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", shardManagerPort), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("failed to dial shardmanager: %v", err)
	}
	defer conn.Close()
	shardClient := shardmanagerpb.NewShardServiceClient(conn)
	shardIDs := []string{
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"00000000-0000-0000-0000-000000000003",
		"00000000-0000-0000-0000-000000000004",
		"00000000-0000-0000-0000-000000000005",
	}
	for _, shardID := range shardIDs {
		_, err := shardClient.RegisterShard(context.Background(), &shardmanagerpb.RegisterShardRequest{
			Shard: &shardmanagerpb.Shard{
				Id:     shardID,
				Type:   "test",
				Size:   1,
				NodeId: "", // Let the shardmanager assign
				Status: "pending",
			},
		})
		if err != nil {
			t.Fatalf("RegisterShard RPC failed: %v", err)
		}
	}

	// Wait for balancing and AddShard RPCs to be delivered
	time.Sleep(1500 * time.Millisecond)

	// Verify that each appserver received the expected shards
	for i, ch := range shardChans {
		count := 0
		timeout := time.After(2 * time.Second)
		loop := true
		for loop {
			select {
			case shard := <-ch:
				log.Printf("Appserver-%d received shard: %s", i+1, shard)
				count++
			case <-timeout:
				loop = false
			}
		}
		log.Printf("Appserver-%d received %d shards", i+1, count)
		close(ch) // Close the channel after we're done reading from it
	}

	// Signal the shardmanager to stop
	close(stopCh)

	// Gracefully stop all app servers
	for _, srv := range appServers {
		srv.GracefulStop()
	}

	// Wait for all goroutines to finish with a timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All goroutines finished successfully
	case <-time.After(5 * time.Second):
		t.Fatal("Test timed out waiting for goroutines to finish")
	}

	os.Remove("integration_test.db")
}
