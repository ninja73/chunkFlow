package storage

import (
	pb "chunkFlow/pkg/proto/storagepb"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type GRPCServer struct {
	pb.UnimplementedStorageServiceServer
	root string
}

func NewGRPCServer(root string) *GRPCServer {
	return &GRPCServer{root: root}
}

func (s *GRPCServer) SaveChunk(ctx context.Context, req *pb.SaveChunkRequest) (*pb.SaveChunkResponse, error) {
	path := filepath.Join(s.root, req.FileId)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(filepath.Join(path, formatChunk(req.ChunkId)), req.Data, 0644)
	if err != nil {
		return nil, err
	}

	return &pb.SaveChunkResponse{
		Ok: true,
	}, nil
}

func (s *GRPCServer) GetChunk(ctx context.Context, req *pb.GetChunkRequest) (*pb.GetChunkResponse, error) {
	path := filepath.Join(s.root, req.FileId, formatChunk(req.ChunkId))
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return &pb.GetChunkResponse{Data: b}, nil
}

func formatChunk(id int32) string {
	return fmt.Sprintf("chunk-%d.bin", id)
}
