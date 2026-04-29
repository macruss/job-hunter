package scraper

import (
	"context"
	"log/slog"

	"github.com/ruslan/job-hunter/internal/domain"
)

// Scraper is the common interface every source must implement.
type Scraper interface {
	Source() domain.Source
	Scrape(ctx context.Context, filter domain.ScrapeFilter) ([]domain.Job, error)
}

// Manager fans out scraping across all registered scrapers.
type Manager struct {
	scrapers []Scraper
	log      *slog.Logger
}

func NewManager(log *slog.Logger, scrapers ...Scraper) *Manager {
	return &Manager{scrapers: scrapers, log: log}
}

func (m *Manager) ScrapeAll(ctx context.Context, filter domain.ScrapeFilter) []domain.Job {
	var all []domain.Job
	for _, s := range m.scrapers {
		jobs, err := s.Scrape(ctx, filter)
		if err != nil {
			m.log.Error("scraper failed", "source", s.Source(), "err", err)
			continue
		}
		m.log.Info("scraped jobs", "source", s.Source(), "count", len(jobs))
		all = append(all, jobs...)
	}
	return all
}
