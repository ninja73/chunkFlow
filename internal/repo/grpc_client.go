package repo

import (
	"chunkFlow/internal/domain"
	pb "chunkFlow/pkg/proto/storagepb"
	"context"
)

type GRPCClient struct {
	client pb.StorageServiceClient
}

func NewGRPCClient(client pb.StorageServiceClient) *GRPCClient {
	return &GRPCClient{
		client: client,
	}
}

func (c *GRPCClient) SaveChunk(ctx context.Context, chunk *domain.FileChunk) error {
	_, err := c.client.SaveChunk(ctx, &pb.SaveChunkRequest{
		FileId:  chunk.FileID,
		ChunkId: int32(chunk.ChunkID),
		Data:    chunk.Data,
	})
	return err
}

func (c *GRPCClient) GetChunk(ctx context.Context, fileID string, chunkID int) (*domain.FileChunk, error) {
	resp, err := c.client.GetChunk(ctx, &pb.GetChunkRequest{
		FileId:  fileID,
		ChunkId: int32(chunkID),
	})
	if err != nil {
		return nil, err
	}

	return &domain.FileChunk{
		FileID:  fileID,
		ChunkID: chunkID,
		Data:    resp.Data,
	}, nil
}
