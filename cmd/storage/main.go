package main

import (
	"chunkFlow/internal/storage"
	pb "chunkFlow/pkg/proto/storagepb"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
)

func main() {
	root := os.Getenv("DATA_DIR")
	if root == "" {
		root = "./data"
	}
	_ = os.MkdirAll(root, 0755)

	srv := grpc.NewServer()
	storageSrv := storage.NewGRPCServer(root)

	pb.RegisterStorageServiceServer(srv, storageSrv)

	go func() {
		lis, err := net.Listen("tcp", ":"+os.Getenv("PORT"))
		if err != nil {
			log.Fatal(err)
		}

		err = srv.Serve(lis)
		if err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	slog.Info("shutting down...")
	srv.GracefulStop()
	slog.Info("server stopped gracefully")
}
