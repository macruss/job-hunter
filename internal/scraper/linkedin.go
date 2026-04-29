package scraper

// LinkedIn scraping notes:
//
// LinkedIn aggressively blocks automated scraping and requires login for most job data.
// Two practical approaches for MVP:
//
//  1. Use the unofficial linkedin-scraper approach with a personal session cookie.
//     Set LINKEDIN_SESSION_COOKIE env var from your browser's li_at cookie.
//
//  2. Use LinkedIn's official Jobs API (requires partner access).
//
// This implementation uses approach 1 — personal session cookie — which works
// for personal job-hunting tooling.

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ruslan/job-hunter/internal/domain"
)

type LinkedInScraper struct {
	client        *http.Client
	sessionCookie string // li_at cookie from browser
	baseURL       string
}

func NewLinkedIn(sessionCookie string) *LinkedInScraper {
	return &LinkedInScraper{
		client:        &http.Client{Timeout: 20 * time.Second},
		sessionCookie: sessionCookie,
		baseURL:       "https://www.linkedin.com",
	}
}

func (l *LinkedInScraper) Source() domain.Source { return domain.SourceLinkedIn }

func (l *LinkedInScraper) Scrape(ctx context.Context, filter domain.ScrapeFilter) ([]domain.Job, error) {
	if l.sessionCookie == "" {
		return nil, fmt.Errorf("linkedin: LINKEDIN_SESSION_COOKIE not set, skipping")
	}

	params := url.Values{}
	params.Set("keywords", strings.Join(filter.Keywords, " "))
	params.Set("start", "0")
	params.Set("count", "25")
	if filter.Remote {
		params.Set("f_WT", "2") // remote filter
	}

	// LinkedIn Voyager API (unofficial, may break on UI updates)
	apiURL := fmt.Sprintf(
		"%s/voyager/api/jobs/jobPostings?%s",
		l.baseURL, params.Encode(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("linkedin build req: %w", err)
	}
	req.Header.Set("Cookie", "li_at="+l.sessionCookie)
	req.Header.Set("Csrf-Token", "ajax:0") // required by Voyager
	req.Header.Set("X-Restli-Protocol-Version", "2.0.0")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("linkedin fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("linkedin: session cookie expired or invalid")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("linkedin: status %d", resp.StatusCode)
	}

	return parseLinkedInResponse(resp)
}

// linkedInJobsResponse mirrors a subset of LinkedIn Voyager response.
type linkedInJobsResponse struct {
	Elements []struct {
		EntityUrn   string `json:"entityUrn"`
		Title       string `json:"title"`
		CompanyName string `json:"companyName"`
		Location    struct {
			DefaultLocalizedName string `json:"defaultLocalizedName"`
		} `json:"formattedLocation"`
		Description struct {
			Text string `json:"text"`
		} `json:"description"`
	} `json:"elements"`
}

func parseLinkedInResponse(resp *http.Response) ([]domain.Job, error) {
	var raw linkedInJobsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("linkedin parse: %w", err)
	}

	jobs := make([]domain.Job, 0, len(raw.Elements))
	for _, el := range raw.Elements {
		// Extract numeric ID from entityUrn like "urn:li:jobPosting:3939393"
		extID := el.EntityUrn
		if i := strings.LastIndex(extID, ":"); i >= 0 {
			extID = extID[i+1:]
		}

		jobs = append(jobs, domain.Job{
			ID:          uuid.New(),
			ExternalID:  extID,
			Source:      domain.SourceLinkedIn,
			Title:       el.Title,
			Company:     el.CompanyName,
			Description: el.Description.Text,
			Location:    el.Location.DefaultLocalizedName,
			URL:         fmt.Sprintf("https://www.linkedin.com/jobs/view/%s/", extID),
			ScrapedAt:   time.Now(),
			CreatedAt:   time.Now(),
		})
	}
	return jobs, nil
}
