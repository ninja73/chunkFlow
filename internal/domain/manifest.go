package domain

type ChunkLocation struct {
	ChunkIndex int `db:"chunk_index"`
	RepoIndex  int `db:"repo_index"`
}

type FileManifest struct {
	FileID    string          `db:"file_id"`
	Completed bool            `db:"completed"`
	Chunks    []ChunkLocation `db:"-"`
}
