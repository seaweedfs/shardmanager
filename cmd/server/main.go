package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/server"
	"google.golang.org/grpc"
)

var (
	port   = flag.Int("port", 7427, "The server port")
	dbConn = flag.String("db", "postgres://postgres:postgres@localhost:5432/shardmanager?sslmode=disable", "Database connection string")
)

func main() {
	flag.Parse()

	// Initialize database connection
	database, err := db.NewDB(*dbConn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Create gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	shardmanagerpb.RegisterNodeServiceServer(s, server.NewServer(database))
	shardmanagerpb.RegisterShardServiceServer(s, server.NewServer(database))
	shardmanagerpb.RegisterPolicyServiceServer(s, server.NewServer(database))
	shardmanagerpb.RegisterMonitoringServiceServer(s, server.NewServer(database))
	shardmanagerpb.RegisterFailureServiceServer(s, server.NewServer(database))

	// Handle graceful shutdown
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	log.Printf("Server listening on port %d", *port)

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down server...")
	s.GracefulStop()
	log.Println("Server stopped")
}
