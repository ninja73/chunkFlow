package repo

import (
	"chunkFlow/internal/domain"
	"context"
	"errors"
	"sync"
)

type InMemoryManifestRepo struct {
	mu    sync.RWMutex
	store map[string]*domain.FileManifest
}

func NewInMemoryManifestRepo() *InMemoryManifestRepo {
	return &InMemoryManifestRepo{store: make(map[string]*domain.FileManifest)}
}

func (r *InMemoryManifestRepo) Create(_ context.Context, fileID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[fileID]; ok {
		return errors.New("manifest exists")
	}

	r.store[fileID] = &domain.FileManifest{FileID: fileID, Chunks: nil}

	return nil
}

func (r *InMemoryManifestRepo) AppendChunk(_ context.Context, fileID string, loc domain.ChunkLocation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	m, ok := r.store[fileID]
	if !ok {
		return errors.New("manifest not found")
	}
	m.Chunks = append(m.Chunks, loc)
	r.store[fileID] = m
	return nil
}

func (r *InMemoryManifestRepo) Get(_ context.Context, fileID string) (*domain.FileManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.store[fileID]
	if !ok {
		return nil, errors.New("not found")
	}
	return m, nil
}

func (r *InMemoryManifestRepo) MarkCompleted(_ context.Context, fileID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.store[fileID]
	if !ok {
		return errors.New("not found")
	}

	m.Completed = true
	r.store[fileID] = m

	return nil
}
