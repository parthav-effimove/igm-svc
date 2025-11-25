CREATE TABLE IF NOT EXISTS ondc_callbacks (
    id BIGSERIAL PRIMARY KEY,

    transaction_id TEXT NOT NULL,
    message_id TEXT NOT NULL,

    
    payload JSONB NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW()
);


CREATE INDEX IF NOT EXISTS idx_ondc_callbacks_tx
    ON ondc_callbacks (transaction_id);

CREATE INDEX IF NOT EXISTS idx_ondc_callbacks_msg
    ON ondc_callbacks (message_id);
