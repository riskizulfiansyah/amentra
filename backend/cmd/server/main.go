package main

import (
	"log/slog"
	"os"

	"amentra/internal/server"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Error("failed to load .env", "error", err)
		os.Exit(1)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	srv := server.New(":8080")
	slog.Info("listening", "addr", ":8080")
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
