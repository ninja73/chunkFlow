package repo

import (
	"chunkFlow/internal/domain"
	"context"

	"github.com/jmoiron/sqlx"
)

type ManifestRepo struct {
	db *sqlx.DB
}

func NewManifestRepo(db *sqlx.DB) *ManifestRepo {
	return &ManifestRepo{db: db}
}

func (r *ManifestRepo) Create(ctx context.Context, fileID string) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO manifests (file_id, completed) VALUES ($1, false)", fileID)
	return err
}

func (r *ManifestRepo) AppendChunk(ctx context.Context, fileID string, loc domain.ChunkLocation) error {
	var m domain.FileManifest
	err := r.db.GetContext(ctx, &m, "SELECT file_id, completed FROM manifests WHERE file_id = $1", fileID)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx,
		"INSERT INTO chunk_locations (file_id, chunk_index, repo_index) VALUES ($1, $2, $3)",
		m.FileID, loc.ChunkIndex, loc.RepoIndex)
	if err != nil {
		return err
	}

	return nil
}

func (r *ManifestRepo) Get(ctx context.Context, fileID string) (*domain.FileManifest, error) {
	var m domain.FileManifest
	err := r.db.GetContext(ctx, &m, "SELECT file_id, completed FROM manifests WHERE file_id = $1", fileID)
	if err != nil {
		return nil, err
	}

	var locations []domain.ChunkLocation
	err = r.db.SelectContext(ctx, &locations, "SELECT chunk_index, repo_index FROM chunk_locations WHERE file_id = $1", fileID)
	if err != nil {
		return nil, err
	}

	m.Chunks = locations

	return &m, nil
}

func (r *ManifestRepo) MarkCompleted(ctx context.Context, fileID string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE manifests SET completed = true WHERE file_id = $1", fileID)
	return err
}
