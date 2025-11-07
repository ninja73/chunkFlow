package domain

type FileChunk struct {
	FileID  string
	ChunkID int
	Data    []byte
}
