.PHONY: dev db stop tidy ui

# Start PostgreSQL only
db:
	docker compose up postgres -d

# Run backend in dev mode (requires db running)
dev:
	go run ./cmd/server

# Run frontend
ui:
	cd frontend && npm install && npm run dev

# Run everything with Docker
up:
	docker compose up --build

# Stop all
stop:
	docker compose down

# Tidy go modules
tidy:
	go mod tidy

# Scrape jobs via API (usage: make scrape KW="Go backend")
scrape:
	curl -s -X POST http://localhost:8080/api/jobs/scrape \
		-H 'Content-Type: application/json' \
		-d '{"keywords":["$(KW)"],"remote":true}' | jq .

# List jobs
jobs:
	curl -s http://localhost:8080/api/jobs | jq '[.[] | {title,company,match_score,source}]'

# List applications
apps:
	curl -s http://localhost:8080/api/applications | jq '[.[] | {status, title: .job.title}]'
