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

type DjinniScraper struct {
	client  *http.Client
	baseURL string
}

func NewDjinni() *DjinniScraper {
	return &DjinniScraper{
		client:  &http.Client{Timeout: 20 * time.Second},
		baseURL: "https://djinni.co",
	}
}

func (d *DjinniScraper) Source() domain.Source { return domain.SourceDjinni }

func (d *DjinniScraper) Scrape(ctx context.Context, filter domain.ScrapeFilter) ([]domain.Job, error) {
	params := url.Values{}
	if len(filter.Keywords) > 0 {
		params.Set("primary_keyword", strings.Join(filter.Keywords, " "))
	}
	if filter.Remote {
		params.Set("employment", "remote")
	}

	reqURL := fmt.Sprintf("%s/jobs/?%s", d.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("djinni build req: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; job-hunter-bot/1.0)")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("djinni fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("djinni: status %d", resp.StatusCode)
	}
	return parseDjinni(resp.Body, d.baseURL)
}

// parseDjinni walks the HTML tree and extracts job-list items.
func parseDjinni(r io.Reader, base string) ([]domain.Job, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var jobs []domain.Job
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && hasClass(n, "job-list-item") {
			jobs = append(jobs, djinniJobFromNode(n, base))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return jobs, nil
}

func djinniJobFromNode(n *html.Node, base string) domain.Job {
	j := domain.Job{
		ID:        uuid.New(),
		Source:    domain.SourceDjinni,
		ScrapedAt: time.Now(),
		CreatedAt: time.Now(),
		Currency:  "USD",
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			case n.Data == "a" && hasClass(n, "job-list-item__link"):
				href := attr(n, "href")
				j.ExternalID = strings.TrimPrefix(href, "/jobs/")
				j.URL = base + href
				j.Title = strings.TrimSpace(textContent(n))
			case hasClass(n, "job-list-item__company"):
				j.Company = strings.TrimSpace(textContent(n))
			case hasClass(n, "location-text"):
				j.Location = strings.TrimSpace(textContent(n))
			case n.Data == "span" && hasClass(n, "tag"):
				j.Tags = append(j.Tags, strings.TrimSpace(textContent(n)))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return j
}

// ── HTML helpers ──────────────────────────────────────────────────────────────

func hasClass(n *html.Node, cls string) bool {
	for _, a := range n.Attr {
		if a.Key == "class" {
			for _, c := range strings.Fields(a.Val) {
				if c == cls {
					return true
				}
			}
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textContent(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}
