package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ruslan/job-hunter/internal/domain"
)

type CVAdapter struct {
	llm LLMClient
}

func NewCVAdapter(llm LLMClient) *CVAdapter {
	return &CVAdapter{llm: llm}
}

// AdaptCV rewrites the CV to highlight skills relevant to the job description.
func (a *CVAdapter) AdaptCV(ctx context.Context, cv domain.CV, job domain.Job) (string, error) {
	prompt := fmt.Sprintf(`You are an expert career coach. Adapt the following CV to better match the job description.

RULES:
- Keep all facts true — do not invent experience
- Reorder and rephrase bullet points to emphasize relevant skills
- Adjust the summary/objective section to match the role
- Keep total length under 600 words
- Return only the adapted CV text, no commentary

JOB TITLE: %s
COMPANY: %s
JOB DESCRIPTION:
%s

ORIGINAL CV:
%s`, job.Title, job.Company, job.Description, cv.Content)

	result, err := a.llm.Complete(ctx, prompt, 2000)
	if err != nil {
		return "", fmt.Errorf("adapt cv: %w", err)
	}
	return result, nil
}

// ScoreMatch returns a 0-100 match score between a CV and job.
func (a *CVAdapter) ScoreMatch(ctx context.Context, cv domain.CV, job domain.Job) (float64, []string, error) {
	prompt := fmt.Sprintf(`Analyze how well this CV matches the job description.

Respond ONLY in this JSON format (no markdown, no extra text):
{
  "score": <integer 0-100>,
  "matching_skills": ["skill1","skill2"],
  "missing_skills": ["skill3","skill4"],
  "summary": "<one sentence>"
}

JOB TITLE: %s
JOB DESCRIPTION:
%s

CV:
%s`, job.Title, job.Description, cv.Content)

	raw, err := a.llm.Complete(ctx, prompt, 500)
	if err != nil {
		return 0, nil, fmt.Errorf("score match: %w", err)
	}

	var resp struct {
		Score         int      `json:"score"`
		MissingSkills []string `json:"missing_skills"`
	}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return 0, nil, fmt.Errorf("score parse: %w (raw: %s)", err, raw)
	}
	return float64(resp.Score), resp.MissingSkills, nil
}

// ParseSkills extracts a list of technical skills from CV text using the local LLM.
func (a *CVAdapter) ParseSkills(ctx context.Context, cvText string) ([]string, error) {
	prompt := `Extract all technical skills from this CV.
Respond ONLY as a JSON array of strings, no markdown, no extra text.
Example: ["Go","PostgreSQL","Docker"]

CV:
` + cvText

	raw, err := a.llm.Complete(ctx, prompt, 300)
	if err != nil {
		return nil, fmt.Errorf("parse skills: %w", err)
	}

	raw = strings.TrimSpace(raw)
	// strip markdown code fences if the model adds them
	if idx := strings.Index(raw, "["); idx > 0 {
		raw = raw[idx:]
	}
	if idx := strings.LastIndex(raw, "]"); idx >= 0 {
		raw = raw[:idx+1]
	}

	var skills []string
	if err := json.Unmarshal([]byte(raw), &skills); err != nil {
		return nil, fmt.Errorf("skills parse: %w (raw: %s)", err, raw)
	}
	return skills, nil
}
