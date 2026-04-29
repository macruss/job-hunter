-- migrations/001_init.sql

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE cvs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_name   VARCHAR(255) NOT NULL,
    content     TEXT NOT NULL,             -- raw text extracted from PDF
    parsed_skills TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE jobs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    external_id VARCHAR(512) NOT NULL,
    source      VARCHAR(50)  NOT NULL,     -- djinni | linkedin | dou | workua
    title       VARCHAR(500) NOT NULL,
    company     VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL,
    url         VARCHAR(1000) NOT NULL,
    location    VARCHAR(255),
    salary_min  INTEGER,
    salary_max  INTEGER,
    currency    VARCHAR(10),
    tags        TEXT[]       NOT NULL DEFAULT '{}',
    match_score FLOAT        NOT NULL DEFAULT 0,
    scraped_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (source, external_id)
);

CREATE TABLE applications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id      UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    cv_id       UUID REFERENCES cvs(id),
    adapted_cv  TEXT,
    status      VARCHAR(50) NOT NULL DEFAULT 'new'
                    CHECK (status IN ('new','applied','screening','interview','offer','rejected')),
    notes       TEXT        NOT NULL DEFAULT '',
    applied_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_source       ON jobs(source);
CREATE INDEX idx_jobs_match_score  ON jobs(match_score DESC);
CREATE INDEX idx_jobs_scraped_at   ON jobs(scraped_at DESC);
CREATE INDEX idx_applications_status  ON applications(status);
CREATE INDEX idx_applications_job_id  ON applications(job_id);
