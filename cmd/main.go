package main

import (
	"api-proxy/internal/api"
	"api-proxy/internal/config"
	"api-proxy/internal/db"
	"log"
	"log/slog"
	"os"
)

func main() {
	appConfig, err := config.LoadConfig("config.yml")

	if err != nil {
		log.Fatal(err)
	}

	slog.SetDefault(
		slog.New(slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: determineLogLevel(appConfig.LoggingConfig.Level)},
		)),
	)

	database, err := db.Connect(appConfig.DB)

	if err != nil {
		slog.Error("Unable to connect to DB", "error", err)
		os.Exit(1)
	}

	slog.Info("successfully connected to database")

	if migrationErr := db.RunMigration(database); migrationErr != nil {
		slog.Error("Unable to run migration", "error", migrationErr)
		os.Exit(1)
	}

	slog.Info("migrations complete")

	server := api.NewServer(appConfig, database)

	slog.Info("Listening on port", "port", appConfig.Server.Port)

	if err = server.Start(); err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
	}

	slog.Info("Server stopped.")
}

func determineLogLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
