ALTER TABLE download_queue ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE download_queue ADD COLUMN last_error TEXT;
