CREATE TABLE IF NOT EXISTS jobs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url             TEXT,
    url_canonical   TEXT,
    title           VARCHAR(500)  NOT NULL,
    title_normalized VARCHAR(500) NOT NULL,
    company         VARCHAR(200)  NOT NULL,
    company_normalized VARCHAR(200) NOT NULL,
    status          VARCHAR(10)   NOT NULL CHECK (status IN ('draft', 'tracked')),
    notes           VARCHAR(2000),
    referrer        VARCHAR(200),
    date_requested  TIMESTAMPTZ   NOT NULL,
    date_added      TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_jobs_url_canonical
    ON jobs (url_canonical)
    WHERE url_canonical IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_jobs_company_title
    ON jobs (company_normalized, title_normalized);

CREATE INDEX IF NOT EXISTS idx_jobs_status
    ON jobs (status);

CREATE INDEX IF NOT EXISTS idx_jobs_date_requested
    ON jobs (date_requested DESC);
