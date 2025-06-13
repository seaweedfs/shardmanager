package server

import (
	"github.com/seaweedfs/shardmanager/db"
	"github.com/seaweedfs/shardmanager/shardmanagerpb"
)

type Server struct {
	shardmanagerpb.UnimplementedNodeServiceServer
	shardmanagerpb.UnimplementedShardServiceServer
	shardmanagerpb.UnimplementedPolicyServiceServer
	shardmanagerpb.UnimplementedMonitoringServiceServer
	shardmanagerpb.UnimplementedFailureServiceServer

	db db.DBOperations
}

func NewServer(db db.DBOperations) *Server {
	return &Server{db: db}
}
