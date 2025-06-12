package server

import (
	"context"
	"encoding/json"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/google/uuid"
	"github.com/seaweedfs/shardmanager/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PolicyService implementation
func (s *Server) SetPolicy(ctx context.Context, req *shardmanagerpb.SetPolicyRequest) (*shardmanagerpb.SetPolicyResponse, error) {
	policy := &db.Policy{
		ID:         uuid.New(),
		PolicyType: req.PolicyType,
		Parameters: json.RawMessage(req.Parameters),
	}

	if err := s.db.SetPolicy(ctx, policy); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.SetPolicyResponse{
		Success: true,
		Message: "Policy set successfully",
	}, nil
}

func (s *Server) GetPolicy(ctx context.Context, req *shardmanagerpb.GetPolicyRequest) (*shardmanagerpb.GetPolicyResponse, error) {
	policy, err := s.db.GetPolicy(ctx, req.PolicyType)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if policy == nil {
		return nil, status.Error(codes.NotFound, "policy not found")
	}

	return &shardmanagerpb.GetPolicyResponse{
		PolicyType: policy.PolicyType,
		Parameters: string(policy.Parameters),
	}, nil
}
