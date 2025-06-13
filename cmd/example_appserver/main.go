package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"
	"google.golang.org/grpc"
)

type appShardServer struct {
	shardmanagerpb.UnimplementedAppShardServiceServer
}

func (s *appShardServer) AddShard(ctx context.Context, req *shardmanagerpb.AddShardRequest) (*shardmanagerpb.AddShardResponse, error) {
	log.Printf("AddShard called: shard_id=%s, role=%s", req.ShardId, req.Role)
	return &shardmanagerpb.AddShardResponse{Success: true, Message: "Shard added"}, nil
}

func (s *appShardServer) DropShard(ctx context.Context, req *shardmanagerpb.DropShardRequest) (*shardmanagerpb.DropShardResponse, error) {
	log.Printf("DropShard called: shard_id=%s", req.ShardId)
	return &shardmanagerpb.DropShardResponse{Success: true, Message: "Shard dropped"}, nil
}

func (s *appShardServer) ChangeRole(ctx context.Context, req *shardmanagerpb.ChangeRoleRequest) (*shardmanagerpb.ChangeRoleResponse, error) {
	log.Printf("ChangeRole called: shard_id=%s, current_role=%s, new_role=%s", req.ShardId, req.CurrentRole, req.NewRole)
	return &shardmanagerpb.ChangeRoleResponse{Success: true, Message: "Role changed"}, nil
}

func (s *appShardServer) PrepareAddShard(ctx context.Context, req *shardmanagerpb.PrepareAddShardRequest) (*shardmanagerpb.PrepareAddShardResponse, error) {
	log.Printf("PrepareAddShard called: shard_id=%s, current_owner=%s, role=%s", req.ShardId, req.CurrentOwner, req.Role)
	return &shardmanagerpb.PrepareAddShardResponse{Success: true, Message: "Prepared to add shard"}, nil
}

func (s *appShardServer) PrepareDropShard(ctx context.Context, req *shardmanagerpb.PrepareDropShardRequest) (*shardmanagerpb.PrepareDropShardResponse, error) {
	log.Printf("PrepareDropShard called: shard_id=%s, new_owner=%s, role=%s", req.ShardId, req.NewOwner, req.Role)
	return &shardmanagerpb.PrepareDropShardResponse{Success: true, Message: "Prepared to drop shard"}, nil
}

func registerWithShardManager(shardManagerAddr, nodeID, appServerAddr string) {
	conn, err := grpc.Dial(shardManagerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to shardmanager: %v", err)
	}
	defer conn.Close()
	client := shardmanagerpb.NewNodeServiceClient(conn)

	req := &shardmanagerpb.RegisterNodeRequest{
		Node: &shardmanagerpb.Node{
			Id:       nodeID,
			Location: appServerAddr,
			Capacity: 100, // example value
			Status:   "active",
		},
	}
	resp, err := client.RegisterNode(context.Background(), req)
	if err != nil || !resp.Success {
		log.Fatalf("RegisterNode failed: %v, message: %s", err, resp.GetMessage())
	}
	log.Printf("Registered with shardmanager: %s", resp.GetMessage())
}

func sendHeartbeats(shardManagerAddr, nodeID string) {
	conn, err := grpc.Dial(shardManagerAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to shardmanager for heartbeat: %v", err)
	}
	client := shardmanagerpb.NewNodeServiceClient(conn)

	go func() {
		for {
			req := &shardmanagerpb.HeartbeatRequest{
				NodeId: nodeID,
				Status: "active",
				Load:   0, // example value
			}
			_, err := client.Heartbeat(context.Background(), req)
			if err != nil {
				log.Printf("Heartbeat failed: %v", err)
			}
			time.Sleep(10 * time.Second)
		}
	}()
}

func main() {
	nodeID := "appserver-1"
	appServerAddr := "localhost:50051"
	shardManagerAddr := "localhost:6000"

	// Register with shardmanager
	registerWithShardManager(shardManagerAddr, nodeID, appServerAddr)
	// Start sending heartbeats
	sendHeartbeats(shardManagerAddr, nodeID)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	shardmanagerpb.RegisterAppShardServiceServer(grpcServer, &appShardServer{})
	log.Println("AppShardService server listening on :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
