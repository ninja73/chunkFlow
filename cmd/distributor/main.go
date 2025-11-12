package main

import (
	"chunkFlow/internal/app"
	"chunkFlow/internal/app/rest"
	"chunkFlow/internal/balancer"
	"chunkFlow/internal/repo"
	"chunkFlow/pkg/proto/storagepb"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	clients, err := InitStoreNodes()
	if err != nil {
		log.Fatal(err)
	}

	db, err := InitDB()
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	manifest := repo.NewManifestRepo(db)
	orch := app.NewOrchestrator(InitRoundRobinBalance(manifest, clients))
	handler := rest.NewHandler(orch)

	r := chi.NewRouter()

	r.Post("/upload", handler.Upload)
	r.Get("/download/{fileID}", handler.Download)

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("listen failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	slog.Info("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}

	slog.Info("server stopped gracefully")
}

func InitRoundRobinBalance(manifestRepo *repo.ManifestRepo, clients []*repo.GRPCClient) *balancer.RoundRobinBalancer {
	clientsStore := make([]balancer.ClientRepo, 0, len(clients))
	for i := range clients {
		clientsStore = append(clientsStore, clients[i])
	}

	return balancer.NewRoundRobinBalancer(manifestRepo, clientsStore)
}

func InitStoreNodes() ([]*repo.GRPCClient, error) {
	raw := os.Getenv("STORAGE_NODES")
	nodes := strings.Split(raw, ",")

	storeClients := make([]*repo.GRPCClient, 0, len(nodes))
	for _, nodeAddr := range nodes {
		storeCln, err := NewGRPCStorageClient(nodeAddr)
		if err != nil {
			return nil, err
		}
		storeClients = append(storeClients, storeCln)
	}

	return storeClients, nil
}

func NewGRPCStorageClient(address string) (*repo.GRPCClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to storage %s: %w", address, err)
	}

	return repo.NewGRPCClient(storagepb.NewStorageServiceClient(conn)), nil
}

func InitDB() (*sqlx.DB, error) {
	db, err := sqlx.Connect("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(15 * time.Second)

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
