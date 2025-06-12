package server

import (
	"context"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MonitoringService implementation
func (s *Server) GetDistribution(ctx context.Context, req *shardmanagerpb.GetDistributionRequest) (*shardmanagerpb.GetDistributionResponse, error) {
	shards, err := s.db.ListShards(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	distribution := make(map[string]*shardmanagerpb.ShardList)
	for _, shard := range shards {
		if shard.NodeID == nil {
			continue
		}
		nodeID := shard.NodeID.String()
		if _, exists := distribution[nodeID]; !exists {
			distribution[nodeID] = &shardmanagerpb.ShardList{
				ShardIds: make([]string, 0),
			}
		}
		distribution[nodeID].ShardIds = append(distribution[nodeID].ShardIds, shard.ID.String())
	}

	return &shardmanagerpb.GetDistributionResponse{
		NodeShards: distribution,
	}, nil
}

func (s *Server) GetHealth(ctx context.Context, req *shardmanagerpb.GetHealthRequest) (*shardmanagerpb.GetHealthResponse, error) {
	// Implement health check logic here
	// For now, return a simple summary
	return &shardmanagerpb.GetHealthResponse{
		Summary: "System is healthy",
	}, nil
}
