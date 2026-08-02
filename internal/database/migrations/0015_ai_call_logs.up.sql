-- Persist every AI call (prompt, raw response, timing, and who triggered
-- it) for the admin "Activity" page, replacing the old logs/ai/*.log files.
CREATE TABLE ai_call_logs (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    provider      TEXT NOT NULL,
    method        TEXT NOT NULL,
    status        TEXT NOT NULL,
    execution_ms  INT NOT NULL DEFAULT 0,
    tokens_used   INT NOT NULL DEFAULT 0,
    prompt        TEXT,
    response      TEXT,
    error_message TEXT,
    user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    user_name     TEXT,
    user_email    TEXT,
    user_role     TEXT
);

CREATE INDEX idx_ai_call_logs_created_at ON ai_call_logs(created_at DESC);
