CREATE TABLE IF NOT EXISTS on_issue_status_responses (
    id SERIAL PRIMARY KEY,

    issue_id TEXT NOT NULL,
    transaction_id TEXT,
    message_id TEXT,

    respondent_actions JSONB,
    resolution_provider JSONB,
    resolution JSONB,

    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_on_issue_status_issue_id
    ON on_issue_status_responses (issue_id);
    

CREATE INDEX IF NOT EXISTS idx_on_issue_status_txn_id
    ON on_issue_status_responses (transaction_id);

CREATE INDEX IF NOT EXISTS idx_on_issue_status_msg_id
    ON on_issue_status_responses (message_id);


CREATE INDEX IF NOT EXISTS idx_on_issue_status_issue_created
    ON on_issue_status_responses (issue_id, created_at DESC);