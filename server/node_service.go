package server

import (
	"context"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NodeService implementation
func (s *Server) RegisterNode(ctx context.Context, req *shardmanagerpb.RegisterNodeRequest) (*shardmanagerpb.RegisterNodeResponse, error) {
	node := &db.Node{
		ID:       uuid.New(),
		Location: req.Node.Location,
		Capacity: req.Node.Capacity,
		Status:   req.Node.Status,
	}

	if err := s.db.RegisterNode(ctx, node); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.RegisterNodeResponse{
		Success: true,
		Message: "Node registered successfully",
	}, nil
}

func (s *Server) Heartbeat(ctx context.Context, req *shardmanagerpb.HeartbeatRequest) (*shardmanagerpb.HeartbeatResponse, error) {
	nodeID, err := uuid.Parse(req.NodeId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid node ID")
	}

	if err := s.db.UpdateNodeHeartbeat(ctx, nodeID, req.Status, req.Load); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.HeartbeatResponse{Success: true}, nil
}

func (s *Server) ListNodes(ctx context.Context, req *shardmanagerpb.ListNodesRequest) (*shardmanagerpb.ListNodesResponse, error) {
	nodes, err := s.db.ListNodes(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	pbNodes := make([]*shardmanagerpb.Node, len(nodes))
	for i, node := range nodes {
		pbNodes[i] = &shardmanagerpb.Node{
			Id:       node.ID.String(),
			Location: node.Location,
			Capacity: node.Capacity,
			Status:   node.Status,
		}
	}

	return &shardmanagerpb.ListNodesResponse{Nodes: pbNodes}, nil
}
