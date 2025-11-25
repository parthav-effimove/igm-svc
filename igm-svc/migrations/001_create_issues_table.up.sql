CREATE TABLE IF NOT EXISTS issues (
    id SERIAL PRIMARY KEY,
    
    -- Core identifiers
    issue_id VARCHAR(255) UNIQUE NOT NULL,
    order_id VARCHAR(255) NOT NULL,
    user_id UUID NOT NULL,
    transaction_id VARCHAR(255),  
    
    -- Network participants
    bpp_id VARCHAR(255),
    bpp_uri TEXT,  
    
    -- User/Complainant info
    user_name VARCHAR(255),
    user_phone VARCHAR(50),
    user_email VARCHAR(255),
    
    -- Issue classification
    category VARCHAR(100),
    sub_category VARCHAR(100),
    issue_type VARCHAR(50),
    status VARCHAR(50),
    respondent_status VARCHAR(50),
    
    -- Description
    description_short TEXT,
    description_long TEXT,
    description_url TEXT,
    description_content_type VARCHAR(100),
    images JSONB,
    
    -- Source
    source_npid VARCHAR(255),
    source_type VARCHAR(100),
    
    -- Expected timelines (ISO 8601 duration strings)
    expected_response_time VARCHAR(50),
    expected_resolution_time VARCHAR(50),
    
    -- Rating (for CLOSE operation)
    rating VARCHAR(50),
    
    -- JSONB fields
    order_details JSONB,
    complainant_actions JSONB,
    respondent_actions JSONB,
    resolution_provider JSONB,
    resolution JSONB,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);


CREATE INDEX idx_issues_issue_id ON issues(issue_id);
CREATE INDEX idx_issues_order_id ON issues(order_id);
CREATE INDEX idx_issues_user_id ON issues(user_id);
CREATE INDEX idx_issues_transaction_id ON issues(transaction_id);
CREATE INDEX idx_issues_bpp_id ON issues(bpp_id);
CREATE INDEX idx_issues_status ON issues(status);
CREATE INDEX idx_issues_deleted_at ON issues(deleted_at);
CREATE INDEX idx_issues_created_at ON issues(created_at);


COMMENT ON TABLE issues IS 'Stores IGM (Issue and Grievance Management) issues from ONDC network';
COMMENT ON COLUMN issues.issue_id IS 'ONDC unique issue identifier (IssueBody.ID)';
COMMENT ON COLUMN issues.transaction_id IS 'ONDC transaction ID linking all related actions';
COMMENT ON COLUMN issues.images IS 'JSON array of image URLs';
COMMENT ON COLUMN issues.order_details IS 'JSON object: {id, provider_id, state, items, fulfillments}';
COMMENT ON COLUMN issues.complainant_actions IS 'JSON array of complainant actions';
COMMENT ON COLUMN issues.respondent_actions IS 'JSON array of respondent (BPP) actions';