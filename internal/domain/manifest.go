package domain

type ChunkLocation struct {
	ChunkIndex int
	RepoIndex  int
	Size       int64
}

type FileManifest struct {
	FileID    string
	Completed bool
	Chunks    []ChunkLocation
}
