package storage

import (
	pb "chunkFlow/pkg/proto/storagepb"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const streamBuf = 32 * 1024

type GRPCServer struct {
	pb.UnimplementedStorageServiceServer
	root string
}

func NewGRPCServer(root string) *GRPCServer {
	return &GRPCServer{root: root}
}

func (s *GRPCServer) UploadChunk(stream pb.StorageService_UploadChunkServer) error {
	var outFile *os.File
	var filePath string
	defer func() {
		if outFile != nil {
			_ = outFile.Close()
		}
	}()

	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.UploadAck{Ok: true, Msg: "ok"})
		}
		if err != nil {
			return err
		}

		if outFile == nil {
			dir := filepath.Join(s.root, msg.FileId)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			filePath = filepath.Join(dir, formatChunk(msg.ChunkIndex))
			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			outFile = f
		}
		if len(msg.Data) > 0 {
			if _, err := outFile.Write(msg.Data); err != nil {
				return err
			}
		}
	}
}

func (s *GRPCServer) DownloadChunk(req *pb.ChunkRequest, stream pb.StorageService_DownloadChunkServer) error {
	filePath := filepath.Join(s.root, req.FileId, formatChunk(req.ChunkIndex))

	f, err := os.Open(filePath)
	if err != nil {
		return status.Errorf(codes.NotFound, "chunk not found")
	}

	defer f.Close()

	buf := make([]byte, streamBuf)

	for {
		n, err := f.Read(buf)
		if n > 0 {
			if err := stream.Send(&pb.ChunkChunk{
				FileId:     req.FileId,
				ChunkIndex: req.ChunkIndex,
				Data:       buf[:n],
			}); err != nil {
				return err
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}

func formatChunk(id int32) string {
	return fmt.Sprintf("chunk-%d.bin", id)
}
