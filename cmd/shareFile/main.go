package main

import (
	"log/slog"
	"os"

	"github.com/clcosta/uniqueShareFile/internal/database"
	"github.com/clcosta/uniqueShareFile/internal/server"
	"github.com/clcosta/uniqueShareFile/pkg/background"
)

func main() {

	db := database.NewDB()
	worker := background.NewWorker()
	worker.Run()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	s := server.NewServer(":8080", logger)
	s.RunServer(db, worker)
}
