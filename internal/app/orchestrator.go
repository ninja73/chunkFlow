package app

import (
	"context"
	"errors"
	"io"
	"log/slog"
)

const bufferSize = 1024 * 1024 * 10 // 10MB

type Balancer interface {
	StartUpload(ctx context.Context) (string, error)
	SaveChunk(ctx context.Context, fileID string, chunkID int, data []byte) error
	GetChunk(ctx context.Context, fileID string, chunkID int) ([]byte, error)
	CompleteUpload(ctx context.Context, fileID string) error
}

type Orchestrator struct {
	dist Balancer
}

func NewOrchestrator(dist Balancer) *Orchestrator {
	return &Orchestrator{
		dist: dist,
	}
}

func (o *Orchestrator) Upload(ctx context.Context, r io.Reader) (string, error) {
	fileID, err := o.dist.StartUpload(ctx)
	if err != nil {
		return "", err
	}

	buf := make([]byte, bufferSize)

	chunkID := 0
	for {
		if ctx.Err() != nil {
			return "", errors.New("context canceled")
		}

		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		err = o.dist.SaveChunk(ctx, fileID, chunkID, buf[:n])
		if err != nil {
			return "", err
		}

		chunkID++
	}

	err = o.dist.CompleteUpload(ctx, fileID)
	if err != nil {
		return "", err
	}

	return fileID, nil
}

func (o *Orchestrator) Download(ctx context.Context, fileID string) (io.Reader, error) {
	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()

		chunkID := 0
		for {
			if ctx.Err() != nil {
				slog.Error("context canceled")
				return
			}

			data, err := o.dist.GetChunk(ctx, fileID, chunkID)
			if err == io.EOF {
				return
			}

			if err != nil {
				slog.Error(err.Error())
				return
			}

			_, err = pw.Write(data)
			if err != nil {
				slog.Error(err.Error())
				return
			}

			chunkID++
		}
	}()

	return pr, nil
}
