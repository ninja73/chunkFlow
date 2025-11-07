package rest

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Orchestrator interface {
	Upload(ctx context.Context, r io.Reader) (string, error)
	Download(ctx context.Context, fileID string) (io.Reader, error)
}

type Handler struct {
	orchestrator Orchestrator
}

func NewHandler(orchestrator Orchestrator) *Handler {
	return &Handler{orchestrator: orchestrator}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer file.Close()

	id, err := h.orchestrator.Upload(r.Context(), file)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, "upload file", http.StatusInternalServerError)
		return
	}

	_, _ = w.Write([]byte(id))
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "fileID")
	reader, err := h.orchestrator.Download(r.Context(), id)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	_, _ = io.Copy(w, reader)
}
