package repo

import (
	"chunkFlow/internal/domain"
	pb "chunkFlow/pkg/proto/storagepb"
	"context"
	"fmt"
	"io"
)

const streamBuf = 32 * 1024

type GRPCClient struct {
	client pb.StorageServiceClient
}

func NewGRPCClient(client pb.StorageServiceClient) *GRPCClient {
	return &GRPCClient{
		client: client,
	}
}

func (c *GRPCClient) UploadChunk(ctx context.Context, chunk *domain.FileChunk, data io.Reader) error {
	stream, err := c.client.UploadChunk(ctx)
	if err != nil {
		return err
	}

	buf := make([]byte, streamBuf)
	for {
		n, err := data.Read(buf)
		if n > 0 {
			req := &pb.ChunkChunk{
				FileId:     chunk.FileID,
				ChunkIndex: int32(chunk.ChunkID),
				Data:       buf[:n],
			}
			if err := stream.Send(req); err != nil {
				_ = stream.CloseSend()
				return err
			}
		}
		if err == io.EOF {
			ack, err := stream.CloseAndRecv()
			if err != nil {
				return err
			}
			if !ack.Ok {
				return fmt.Errorf("upload ack not ok: %s", ack.Msg)
			}
			return nil
		}

		if err != nil {
			_ = stream.CloseSend()
			return err
		}
	}
}

func (c *GRPCClient) GetChunk(ctx context.Context, fileID string, chunkID int, w io.Writer) error {
	stream, err := c.client.DownloadChunk(ctx, &pb.ChunkRequest{
		FileId:     fileID,
		ChunkIndex: int32(chunkID),
	})
	if err != nil {
		return err
	}

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if len(msg.Data) > 0 {
			_, err = w.Write(msg.Data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
