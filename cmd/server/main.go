package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ruslan/job-hunter/internal/adapter"
	"github.com/ruslan/job-hunter/internal/api"
	"github.com/ruslan/job-hunter/internal/scraper"
	"github.com/ruslan/job-hunter/internal/storage"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/jobhunter?sslmode=disable"
	}

	db, err := storage.Connect(dsn)
	if err != nil {
		log.Error("db connect", "err", err)
		os.Exit(1)
	}
	if err := db.RunMigrations("migrations/001_init.sql"); err != nil {
		log.Error("migrations", "err", err)
		os.Exit(1)
	}

	cvAdapter := adapter.NewCVAdapter(adapter.NewOllamaClient())

	scraperMgr := scraper.NewManager(log,
		scraper.NewDjinni(),
		scraper.NewLinkedIn(os.Getenv("LINKEDIN_SESSION_COOKIE")),
		scraper.NewDOU(),
		scraper.NewWorkUA(),
	)

	srv := api.NewServer(
		storage.NewJobRepo(db),
		storage.NewApplicationRepo(db),
		storage.NewCVRepo(db),
		scraperMgr,
		cvAdapter,
		log,
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info("starting", "port", port)
	if err := http.ListenAndServe(":"+port, srv); err != nil {
		log.Error("server", "err", err)
		os.Exit(1)
	}
}
