package server

import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/shardmanagerpb"
	"google.golang.org/grpc"
)

// StartShardManagerServer starts the real shardmanager gRPC server.
// It blocks until stopCh is closed.
func StartShardManagerServer(dbConnStr string, grpcAddr string, stopCh <-chan struct{}) error {
	log.Printf("[DEBUG] StartShardManagerServer called with dbConnStr: %s", dbConnStr)
	// Detect driver from dbConnStr
	var database *db.DB
	var err error
	if strings.HasPrefix(dbConnStr, "file:") {
		log.Printf("[DEBUG] Using sqlite3 driver for dbConnStr: %s", dbConnStr)
		database, err = db.NewDBWithDriver("sqlite3", dbConnStr)
		if err == nil {
			err = db.InitSQLiteSchema(database)
			if err != nil {
				log.Printf("[ERROR] Failed to initialize SQLite schema: %v", err)
				return fmt.Errorf("failed to initialize SQLite schema: %w", err)
			}
		}
	} else {
		log.Printf("[DEBUG] Using postgres driver for dbConnStr: %s", dbConnStr)
		database, err = db.NewDB(dbConnStr)
	}
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer database.Close()

	// Create gRPC server
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s := grpc.NewServer()
	srv := NewServer(database)
	shardmanagerpb.RegisterNodeServiceServer(s, srv)
	shardmanagerpb.RegisterShardServiceServer(s, srv)
	shardmanagerpb.RegisterPolicyServiceServer(s, srv)
	shardmanagerpb.RegisterMonitoringServiceServer(s, srv)
	shardmanagerpb.RegisterFailureServiceServer(s, srv)

	// Start serving in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
	log.Printf("ShardManager server listening on %s", grpcAddr)

	// Wait for stop signal
	<-stopCh
	log.Println("Shutting down ShardManager server...")
	s.GracefulStop()
	log.Println("ShardManager server stopped")
	return nil
}
