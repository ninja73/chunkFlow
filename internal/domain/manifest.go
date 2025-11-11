package domain

type ChunkLocation struct {
	ChunkIndex int
	RepoIndex  int
}

type FileManifest struct {
	FileID    string
	Completed bool
	Chunks    []ChunkLocation
}
