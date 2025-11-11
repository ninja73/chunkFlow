package balancer

import (
	"chunkFlow/internal/domain"
	"context"
	"errors"
	"io"
	"sync"

	"github.com/google/uuid"
)

type ClientRepo interface {
	UploadChunk(ctx context.Context, chunk *domain.FileChunk, data io.Reader) error
	GetChunk(ctx context.Context, fileID string, chunkID int, pw io.Writer) error
}

type ManifestRepo interface {
	Get(ctx context.Context, fileID string) (*domain.FileManifest, error)
	Create(ctx context.Context, fileID string) error
	MarkCompleted(ctx context.Context, fileID string) error
	AppendChunk(ctx context.Context, fileID string, loc domain.ChunkLocation) error
}

type RoundRobinBalancer struct {
	clients      []ClientRepo
	manifestRepo ManifestRepo
	mu           sync.Mutex
	next         int
}

func NewRoundRobinBalancer(manifestRepo ManifestRepo, clients []ClientRepo) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		clients:      clients,
		manifestRepo: manifestRepo,
	}
}

func (d *RoundRobinBalancer) StartUpload(ctx context.Context) (string, error) {
	fileID := uuid.New().String()
	err := d.manifestRepo.Create(ctx, fileID)
	if err != nil {
		return "", err
	}

	return fileID, nil
}

func (d *RoundRobinBalancer) CompleteUpload(ctx context.Context, fileID string) error {
	return d.manifestRepo.MarkCompleted(ctx, fileID)
}

func (d *RoundRobinBalancer) SaveChunk(ctx context.Context, fileID string, chunkID int, data io.Reader) error {
	d.mu.Lock()
	targetIndex := d.next
	d.next = (d.next + 1) % len(d.clients)
	d.mu.Unlock()

	chunk := &domain.FileChunk{
		FileID:  fileID,
		ChunkID: chunkID,
	}

	err := d.clients[targetIndex].UploadChunk(ctx, chunk, data)
	if err != nil {
		return err
	}

	loc := domain.ChunkLocation{
		ChunkIndex: chunkID,
		RepoIndex:  targetIndex,
	}

	if err := d.manifestRepo.AppendChunk(ctx, fileID, loc); err != nil {
		return err
	}

	return nil
}

func (d *RoundRobinBalancer) GetChunk(ctx context.Context, fileID string, chunkID int, w io.Writer) error {
	manifest, err := d.manifestRepo.Get(ctx, fileID)
	if err != nil {
		return err
	}

	if !manifest.Completed {
		return errors.New("file not completed")
	}

	if chunkID < 0 || chunkID >= len(manifest.Chunks) {
		return io.EOF
	}
	loc := manifest.Chunks[chunkID]
	repoIdx := loc.RepoIndex
	if repoIdx < 0 || repoIdx >= len(d.clients) {
		return errors.New("invalid repo index in manifest")
	}

	err = d.clients[repoIdx].GetChunk(ctx, fileID, chunkID, w)
	if err != nil {
		return err
	}

	return nil
}

func (d *RoundRobinBalancer) Size() int {
	return len(d.clients)
}
