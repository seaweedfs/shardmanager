package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/seaweedfs/shardmanager/server"
)

var (
	port   = flag.Int("port", 7427, "The server port")
	dbConn = flag.String("db", "postgres://postgres:postgres@localhost:5432/shardmanager?sslmode=disable", "Database connection string")
)

func main() {
	flag.Parse()

	stopCh := make(chan struct{})

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Println("Shutting down server...")
		close(stopCh)
	}()

	grpcAddr := fmt.Sprintf(":%d", *port)
	if err := server.StartShardManagerServer(*dbConn, grpcAddr, stopCh); err != nil {
		log.Fatalf("Failed to start ShardManager server: %v", err)
	}

	log.Println("Server stopped")
}
