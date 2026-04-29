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

type DOUScraper struct {
	client  *http.Client
	baseURL string
}

func NewDOU() *DOUScraper {
	return &DOUScraper{
		client:  &http.Client{Timeout: 20 * time.Second},
		baseURL: "https://jobs.dou.ua",
	}
}

func (d *DOUScraper) Source() domain.Source { return domain.SourceDOU }

func (d *DOUScraper) Scrape(ctx context.Context, filter domain.ScrapeFilter) ([]domain.Job, error) {
	params := url.Values{}
	if len(filter.Keywords) > 0 {
		params.Set("search", strings.Join(filter.Keywords, " "))
	}
	if filter.Remote {
		params.Set("remote", "1")
	}

	reqURL := fmt.Sprintf("%s/vacancies/?%s", d.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("dou build req: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; job-hunter-bot/1.0)")
	req.Header.Set("Referer", d.baseURL)

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("dou fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("dou: status %d", resp.StatusCode)
	}
	return parseDOU(resp.Body, d.baseURL)
}

func parseDOU(r io.Reader, base string) ([]domain.Job, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var jobs []domain.Job
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "li" && hasClass(n, "l-vacancy") {
			jobs = append(jobs, douJobFromNode(n, base))
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return jobs, nil
}

func douJobFromNode(n *html.Node, base string) domain.Job {
	j := domain.Job{
		ID:        uuid.New(),
		Source:    domain.SourceDOU,
		ScrapedAt: time.Now(),
		CreatedAt: time.Now(),
		Currency:  "USD",
	}

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch {
			case n.Data == "a" && hasClass(n, "vt"):
				j.URL = attr(n, "href")
				j.Title = strings.TrimSpace(textContent(n))
				// DOU uses full URL
				if strings.HasPrefix(j.URL, "/") {
					j.URL = base + j.URL
				}
				j.ExternalID = j.URL // DOU vacancy URL is stable
			case hasClass(n, "company"):
				j.Company = strings.TrimSpace(textContent(n))
			case hasClass(n, "cities"):
				j.Location = strings.TrimSpace(textContent(n))
			case n.Data == "span" && hasClass(n, "salary"):
				// parse salary text: e.g. "$3000–$5000"
				parseDOUSalary(&j, strings.TrimSpace(textContent(n)))
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return j
}

func parseDOUSalary(j *domain.Job, s string) {
	// Very simple parser — improve as needed
	s = strings.ReplaceAll(s, "$", "")
	s = strings.ReplaceAll(s, "–", "-")
	parts := strings.Split(s, "-")
	if len(parts) == 2 {
		var min, max int
		fmt.Sscan(strings.TrimSpace(parts[0]), &min)
		fmt.Sscan(strings.TrimSpace(parts[1]), &max)
		j.SalaryMin = &min
		j.SalaryMax = &max
		j.Currency = "USD"
	}
}
