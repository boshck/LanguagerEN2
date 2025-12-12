-- Additional performance indexes

-- Composite index for user queries with date filtering
CREATE INDEX IF NOT EXISTS idx_words_user_date ON words(user_id, created_at DESC);

-- Index for date-based pagination queries
CREATE INDEX IF NOT EXISTS idx_words_date_id ON words(created_at DESC, id);

-- Comment for future reference
COMMENT ON INDEX idx_words_user_date IS 'Optimizes queries filtering by user and ordering by date';
COMMENT ON INDEX idx_words_date_id IS 'Optimizes pagination queries with date ordering';

