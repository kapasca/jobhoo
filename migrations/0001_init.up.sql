-- JOBHOO initial schema

CREATE TYPE user_role AS ENUM ('candidate', 'recruiter', 'super_admin');
CREATE TYPE recruiter_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE job_status AS ENUM ('open', 'closed');
CREATE TYPE employment_type AS ENUM ('full_time', 'part_time', 'contract', 'internship', 'freelance');
CREATE TYPE work_arrangement AS ENUM ('onsite', 'remote', 'hybrid');
CREATE TYPE application_stage AS ENUM ('applied', 'resume_reviewed', 'interview', 'offered', 'hired', 'rejected');
CREATE TYPE application_final_status AS ENUM ('proceed_to_next_stage', 'resume_not_reviewed', 'not_matched');

CREATE TABLE users (
    id              BIGSERIAL PRIMARY KEY,
    email           VARCHAR(255) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    role            user_role NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE candidate_profiles (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    full_name       VARCHAR(255) NOT NULL DEFAULT '',
    resume_path     VARCHAR(500) NOT NULL,
    resume_filename VARCHAR(255) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE recruiter_profiles (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name        VARCHAR(255) NOT NULL DEFAULT '',
    document_path       VARCHAR(500) NOT NULL,
    document_filename   VARCHAR(255) NOT NULL,
    status              recruiter_status NOT NULL DEFAULT 'pending',
    reviewed_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE jobs (
    id                  BIGSERIAL PRIMARY KEY,
    recruiter_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title               VARCHAR(255) NOT NULL,
    position            VARCHAR(255) NOT NULL,
    employment_type     employment_type NOT NULL,
    work_arrangement    work_arrangement NOT NULL,
    location            VARCHAR(255) NOT NULL,
    salary_min          BIGINT,
    salary_max          BIGINT,
    benefits            TEXT NOT NULL DEFAULT '',
    requirements        TEXT NOT NULL DEFAULT '',
    description         TEXT NOT NULL DEFAULT '',
    closing_date        DATE NOT NULL,
    status              job_status NOT NULL DEFAULT 'open',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_recruiter ON jobs(recruiter_id);

CREATE TABLE applications (
    id                  BIGSERIAL PRIMARY KEY,
    job_id              BIGINT NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    candidate_id        BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stage               application_stage NOT NULL DEFAULT 'applied',
    final_status        application_final_status,
    match_score         INTEGER,
    skill_match         INTEGER,
    experience_match    INTEGER,
    education_match     INTEGER,
    resume_path         VARCHAR(500) NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(job_id, candidate_id)
);

CREATE INDEX idx_applications_job ON applications(job_id);
CREATE INDEX idx_applications_candidate ON applications(candidate_id);
CREATE INDEX idx_applications_stage ON applications(stage);

CREATE TABLE sessions (
    id              VARCHAR(64) PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_sessions_user ON sessions(user_id);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);
