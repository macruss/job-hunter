package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ruslan/job-hunter/internal/domain"
	"golang.org/x/net/html"
)

type WorkUAScraper struct {
	client  *http.Client
	baseURL string
}

func NewWorkUA() *WorkUAScraper {
	return &WorkUAScraper{
		client:  &http.Client{Timeout: 20 * time.Second},
		baseURL: "https://www.work.ua",
	}
}

func (w *WorkUAScraper) Source() domain.Source { return domain.SourceWorkUA }

func (w *WorkUAScraper) Scrape(ctx context.Context, filter domain.ScrapeFilter) ([]domain.Job, error) {
	params := url.Values{}
	if len(filter.Keywords) > 0 {
		params.Set("q", strings.Join(filter.Keywords, " "))
	}
	if filter.Remote {
		params.Set("wt", "5") // work.ua: remote type
	}

	reqURL := fmt.Sprintf("%s/jobs/?%s", w.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("workua build req: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; job-hunter-bot/1.0)")

	resp, err := w.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workua fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workua: status %d", resp.StatusCode)
	}
	return parseWorkUA(resp.Body, w.baseURL)
}

func parseWorkUA(r io.Reader, base string) ([]domain.Job, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var jobs []domain.Job
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		// work.ua renders each vacancy in a <div class="card ..."> with data-id
		if n.Type == html.ElementNode && n.Data == "div" &&
			hasClass(n, "card") && attr(n, "data-id") != "" {
			jobs = append(jobs, workuaJobFromNode(n, base))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return jobs, nil
}

func workuaJobFromNode(n *html.Node, base string) domain.Job {
	j := domain.Job{
		ID:         uuid.New(),
		ExternalID: attr(n, "data-id"),
		Source:     domain.SourceWorkUA,
		ScrapedAt:  time.Now(),
		CreatedAt:  time.Now(),
		Currency:   "UAH",
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			case n.Data == "h2" && hasClass(n, "cut-top"):
				if a := firstChild(n, "a"); a != nil {
					href := attr(a, "href")
					j.URL = base + href
					j.Title = strings.TrimSpace(textContent(a))
				}
			case hasClass(n, "add-top-xs") && n.Data == "span":
				j.Company = strings.TrimSpace(textContent(n))
			case hasClass(n, "location-text"):
				j.Location = strings.TrimSpace(textContent(n))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return j
}

func firstChild(n *html.Node, tag string) *html.Node {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == tag {
			return c
		}
	}
	return nil
}
