-- Companies must be approved by an admin before their recruiter can post
-- jobs. This gates the business entity (company), not the user account —
-- a recruiter can sign up and log in immediately, but their company sits
-- in 'pending' until an admin reviews it. This matches how the product
-- actually works: the thing being vetted is "is this a real company we
-- want listing jobs on JOBHOO", not "is this person allowed to have an
-- account".

CREATE TYPE company_status AS ENUM ('pending', 'approved', 'rejected');

ALTER TABLE companies ADD COLUMN status company_status NOT NULL DEFAULT 'pending';
ALTER TABLE companies ADD COLUMN approved_at TIMESTAMPTZ;
ALTER TABLE companies ADD COLUMN approved_by UUID REFERENCES users(id);
ALTER TABLE companies ADD COLUMN rejection_reason TEXT;

CREATE INDEX idx_companies_status ON companies (status);
