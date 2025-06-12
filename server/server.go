package server

import (
	"github.com/seaweedfs/shardmanager/shardmanagerpb"

	"github.com/seaweedfs/shardmanager/server/testutil"
)

type Server struct {
	shardmanagerpb.UnimplementedNodeServiceServer
	shardmanagerpb.UnimplementedShardServiceServer
	shardmanagerpb.UnimplementedPolicyServiceServer
	shardmanagerpb.UnimplementedMonitoringServiceServer
	shardmanagerpb.UnimplementedFailureServiceServer

	db testutil.DBOperations
}

func NewServer(db testutil.DBOperations) *Server {
	return &Server{db: db}
}
