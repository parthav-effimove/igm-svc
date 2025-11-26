-- Drop indexes first
DROP INDEX IF EXISTS idx_on_issue_status_issue_created;
DROP INDEX IF EXISTS idx_on_issue_status_msg_id;
DROP INDEX IF EXISTS idx_on_issue_status_txn_id;
DROP INDEX IF EXISTS idx_on_issue_status_issue_id;

-- Then drop the table
DROP TABLE IF EXISTS on_issue_status_responses;