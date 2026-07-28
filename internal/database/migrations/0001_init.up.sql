-- JOBHOO core schema
-- Covers: accounts/auth, companies, jobs, candidate profiles, applications,
-- ATS pipeline stages, saved jobs, and AI-generated artifacts (kept separate
-- from source-of-truth tables so AI output is always clearly provenance-tagged
-- and never silently overwrites human-entered data).

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TYPE user_role AS ENUM ('candidate', 'recruiter', 'admin');

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL UNIQUE,
    password_hash   TEXT NOT NULL,
    role            user_role NOT NULL,
    full_name       TEXT NOT NULL,
    avatar_url      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE companies (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    logo_url        TEXT,
    website         TEXT,
    description     TEXT,
    industry        TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'internship', 'freelance');
CREATE TYPE work_arrangement AS ENUM ('onsite', 'hybrid', 'remote');
CREATE TYPE job_status AS ENUM ('draft', 'published', 'closed', 'archived');

CREATE TABLE jobs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id          UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    created_by          UUID NOT NULL REFERENCES users(id),
    title               TEXT NOT NULL,
    description         TEXT NOT NULL,
    location            TEXT,
    employment_type     employment_type NOT NULL,
    work_arrangement    work_arrangement NOT NULL,
    salary_min          INTEGER,
    salary_max          INTEGER,
    salary_currency     TEXT DEFAULT 'USD',
    must_have_skills    TEXT[] NOT NULL DEFAULT '{}',
    nice_to_have_skills TEXT[] NOT NULL DEFAULT '{}',
    seniority           TEXT,
    status              job_status NOT NULL DEFAULT 'draft',
    published_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_jobs_status_published ON jobs (status, published_at DESC);
CREATE INDEX idx_jobs_company ON jobs (company_id);

CREATE TABLE candidate_profiles (
    user_id         UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    headline        TEXT,
    resume_text     TEXT,
    resume_file_url TEXT,
    skills          TEXT[] NOT NULL DEFAULT '{}',
    experience      JSONB NOT NULL DEFAULT '[]',
    location        TEXT,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE application_stage AS ENUM (
    'applied', 'screening', 'interview', 'offer', 'hired', 'rejected'
);

CREATE TABLE applications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    candidate_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stage           application_stage NOT NULL DEFAULT 'applied',
    cover_note      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (job_id, candidate_id)
);

CREATE INDEX idx_applications_job_stage ON applications (job_id, stage);
CREATE INDEX idx_applications_candidate ON applications (candidate_id);

CREATE TABLE saved_jobs (
    candidate_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (candidate_id, job_id)
);

-- AI-generated artifacts are stored separately from source data so it is
-- always clear what a human entered versus what a model suggested, and so
-- artifacts can be regenerated/expired without touching applications or
-- profiles directly.
CREATE TABLE ai_match_insights (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id          UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    candidate_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,
    score           NUMERIC(5,2),
    summary         TEXT,
    strengths       TEXT[] NOT NULL DEFAULT '{}',
    gaps            TEXT[] NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (job_id, candidate_id)
);
