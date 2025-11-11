package rest

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
)

type Orchestrator interface {
	Upload(ctx context.Context, r io.Reader, totalSize int64) (string, error)
	Download(ctx context.Context, fileID string) (io.Reader, error)
}

type Handler struct {
	orchestrator Orchestrator
}

func NewHandler(orchestrator Orchestrator) *Handler {
	return &Handler{orchestrator: orchestrator}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	var id string
	total := header.Size
	if total <= 0 {
		tmp, err := os.CreateTemp("", "upload-*")
		if err != nil {
			http.Error(w, "create temp file failed", http.StatusInternalServerError)
			return
		}

		defer func() {
			_ = tmp.Close()
			_ = os.Remove(tmp.Name())
		}()

		if _, err := io.Copy(tmp, file); err != nil {
			http.Error(w, "read body failed", http.StatusInternalServerError)
			return
		}

		st, err := tmp.Stat()
		if err != nil {
			http.Error(w, "stat failed", http.StatusInternalServerError)
			return
		}

		total = st.Size()
		if _, err := tmp.Seek(0, io.SeekStart); err != nil {
			http.Error(w, "seek failed", http.StatusInternalServerError)
			return
		}

		id, err = h.orchestrator.Upload(r.Context(), tmp, total)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "upload file", http.StatusInternalServerError)
			return
		}
	} else {
		id, err = h.orchestrator.Upload(r.Context(), file, total)
		if err != nil {
			slog.Error(err.Error())
			http.Error(w, "upload file", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
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
