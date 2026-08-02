-- Add overall_note to store the AI's free-text match explanation (Explain
-- Match feature), separate from summary (ranking rationale) and
-- strengths/gaps (already existing columns).

ALTER TABLE ai_match_insights ADD COLUMN overall_note TEXT;
