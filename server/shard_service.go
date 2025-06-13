package server

import (
	"context"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"log"
	"time"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ShardService implementation
func (s *Server) RegisterShard(ctx context.Context, req *shardmanagerpb.RegisterShardRequest) (*shardmanagerpb.RegisterShardResponse, error) {
	// Use the provided shard ID as a UUID
	shardID, err := uuid.Parse(req.Shard.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shard ID")
	}
	shard := &db.Shard{
		ID:     shardID,
		Type:   req.Shard.Type,
		Size:   req.Shard.Size,
		Status: req.Shard.Status,
	}

	// If no node is specified, find an available node
	if req.Shard.NodeId == "" {
		nodes, err := s.db.ListNodes(ctx)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		if len(nodes) == 0 {
			return nil, status.Error(codes.FailedPrecondition, "no nodes available for shard assignment")
		}

		// Simple round-robin: find the node with the least number of shards
		var selectedNode *db.Node
		minShards := -1
		for _, node := range nodes {
			if node.Status != "active" {
				continue
			}
			shards, err := s.db.ListShards(ctx)
			if err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			nodeShardCount := 0
			for _, shard := range shards {
				if shard.NodeID != nil && *shard.NodeID == node.ID {
					nodeShardCount++
				}
			}
			if minShards == -1 || nodeShardCount < minShards {
				minShards = nodeShardCount
				selectedNode = node
			}
		}

		if selectedNode == nil {
			return nil, status.Error(codes.FailedPrecondition, "no active nodes available for shard assignment")
		}

		shard.NodeID = &selectedNode.ID
	} else {
		nodeID, err := uuid.Parse(req.Shard.NodeId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid node ID")
		}
		shard.NodeID = &nodeID
	}

	if err := s.db.RegisterShard(ctx, shard); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Notify appserver if assigned
	if shard.NodeID != nil {
		go s.notifyAppServerAddShard(context.Background(), *shard.NodeID, shard.ID, "primary")
	}

	return &shardmanagerpb.RegisterShardResponse{
		Success: true,
		Message: "Shard registered successfully",
	}, nil
}

func (s *Server) ListShards(ctx context.Context, req *shardmanagerpb.ListShardsRequest) (*shardmanagerpb.ListShardsResponse, error) {
	shards, err := s.db.ListShards(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbShards := make([]*shardmanagerpb.Shard, len(shards))
	for i, shard := range shards {
		pbShard := &shardmanagerpb.Shard{
			Id:     shard.ID.String(),
			Type:   shard.Type,
			Size:   shard.Size,
			Status: shard.Status,
		}
		if shard.NodeID != nil {
			pbShard.NodeId = shard.NodeID.String()
		}
		pbShards[i] = pbShard
	}

	return &shardmanagerpb.ListShardsResponse{Shards: pbShards}, nil
}

func (s *Server) GetShardInfo(ctx context.Context, req *shardmanagerpb.GetShardInfoRequest) (*shardmanagerpb.GetShardInfoResponse, error) {
	shardID, err := uuid.Parse(req.ShardId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shard ID")
	}

	shard, err := s.db.GetShardInfo(ctx, shardID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if shard == nil {
		return nil, status.Error(codes.NotFound, "shard not found")
	}

	pbShard := &shardmanagerpb.Shard{
		Id:     shard.ID.String(),
		Type:   shard.Type,
		Size:   shard.Size,
		Status: shard.Status,
	}
	if shard.NodeID != nil {
		pbShard.NodeId = shard.NodeID.String()
	}

	return &shardmanagerpb.GetShardInfoResponse{Shard: pbShard}, nil
}

func (s *Server) AssignShard(ctx context.Context, req *shardmanagerpb.AssignShardRequest) (*shardmanagerpb.AssignShardResponse, error) {
	shardID, err := uuid.Parse(req.ShardId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shard ID")
	}

	nodeID, err := uuid.Parse(req.NodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node ID")
	}

	if err := s.db.AssignShard(ctx, shardID, nodeID); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Notify appserver
	go s.notifyAppServerAddShard(context.Background(), nodeID, shardID, "primary")

	return &shardmanagerpb.AssignShardResponse{
		Success: true,
		Message: "Shard assigned successfully",
	}, nil
}

func (s *Server) MigrateShard(ctx context.Context, req *shardmanagerpb.MigrateShardRequest) (*shardmanagerpb.MigrateShardResponse, error) {
	shardID, err := uuid.Parse(req.ShardId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shard ID")
	}

	fromNodeID, err := uuid.Parse(req.FromNodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid source node ID")
	}

	toNodeID, err := uuid.Parse(req.ToNodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid destination node ID")
	}

	// First verify the shard is on the source node
	shard, err := s.db.GetShardInfo(ctx, shardID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if shard == nil {
		return nil, status.Error(codes.NotFound, "shard not found")
	}
	if shard.NodeID == nil || *shard.NodeID != fromNodeID {
		return nil, status.Error(codes.FailedPrecondition, "shard is not on the source node")
	}

	// Update shard status to migrating
	if err := s.db.UpdateShardStatus(ctx, shardID, "migrating"); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Assign shard to new node
	if err := s.db.AssignShard(ctx, shardID, toNodeID); err != nil {
		// Revert status if assignment fails
		_ = s.db.UpdateShardStatus(ctx, shardID, "active")
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Update shard status back to active
	if err := s.db.UpdateShardStatus(ctx, shardID, "active"); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.MigrateShardResponse{
		Success: true,
		Message: "Shard migrated successfully",
	}, nil
}

func (s *Server) UpdateShardStatus(ctx context.Context, req *shardmanagerpb.UpdateShardStatusRequest) (*shardmanagerpb.UpdateShardStatusResponse, error) {
	shardID, err := uuid.Parse(req.ShardId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid shard ID")
	}

	if err := s.db.UpdateShardStatus(ctx, shardID, req.Status); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.UpdateShardStatusResponse{
		Success: true,
		Message: "Shard status updated successfully",
	}, nil
}

// notifyAppServerAddShard contacts the appserver and calls AddShard RPC
func (s *Server) notifyAppServerAddShard(ctx context.Context, nodeID uuid.UUID, shardID uuid.UUID, role string) {
	node, err := s.db.GetNodeInfo(ctx, nodeID)
	if err != nil || node == nil {
		log.Printf("[WARN] Could not find node %s for AddShard notification: %v", nodeID, err)
		return
	}
	conn, err := grpc.Dial(node.Location, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2*time.Second))
	if err != nil {
		log.Printf("[WARN] Could not connect to appserver at %s: %v", node.Location, err)
		return
	}
	defer conn.Close()
	client := shardmanagerpb.NewAppShardServiceClient(conn)
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, err = client.AddShard(ctx, &shardmanagerpb.AddShardRequest{
		ShardId: shardID.String(),
		Role:    role,
	})
	if err != nil {
		log.Printf("[WARN] AddShard RPC to appserver %s failed: %v", node.Location, err)
	}
}
