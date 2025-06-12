package server

import (
	"context"
	"encoding/json"

	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// FailureService implementation
func (s *Server) ReportFailure(ctx context.Context, req *shardmanagerpb.ReportFailureRequest) (*shardmanagerpb.ReportFailureResponse, error) {
	entityID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid entity ID")
	}

	details := json.RawMessage(req.Details)
	if err := s.db.ReportFailure(ctx, req.Type, entityID, details); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &shardmanagerpb.ReportFailureResponse{
		Success: true,
		Message: "Failure reported successfully",
	}, nil
}
