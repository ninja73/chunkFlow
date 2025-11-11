package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
)

type Balancer interface {
	StartUpload(ctx context.Context) (string, error)
	SaveChunk(ctx context.Context, fileID string, chunkID int, data io.Reader) error
	GetChunk(ctx context.Context, fileID string, chunkID int, w io.Writer) error
	CompleteUpload(ctx context.Context, fileID string) error
	Size() int
}

type Orchestrator struct {
	dist Balancer
}

func NewOrchestrator(dist Balancer) *Orchestrator {
	return &Orchestrator{
		dist: dist,
	}
}

func (o *Orchestrator) Upload(ctx context.Context, r io.Reader, totalSize int64) (string, error) {
	fileID, err := o.dist.StartUpload(ctx)
	if err != nil {
		return "", err
	}

	partSize := totalSize / int64(o.dist.Size())
	remainder := totalSize % int64(o.dist.Size())

	for chunkID := 0; chunkID < o.dist.Size(); chunkID++ {
		if ctx.Err() != nil {
			return "", errors.New("context canceled")
		}

		size := partSize
		if chunkID == o.dist.Size()-1 {
			size += remainder
		}

		if size == 0 {
			continue
		}

		lim := io.LimitReader(r, size)

		if err := o.dist.SaveChunk(ctx, fileID, chunkID, lim); err != nil {
			return "", err
		}
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
		for chunkID := 0; chunkID < o.dist.Size(); chunkID++ {
			if ctx.Err() != nil {
				slog.Error("context canceled")
				return
			}

			fmt.Println(fileID, chunkID)

			err := o.dist.GetChunk(ctx, fileID, chunkID, pw)
			if err == io.EOF {
				return
			}

			if err != nil {
				slog.Error(err.Error())
				return
			}
		}
	}()

	return pr, nil
}
