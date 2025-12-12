-- Add fields for hiding words from random pair

-- Add hidden_until field for temporary hiding (7 days)
ALTER TABLE words ADD COLUMN IF NOT EXISTS hidden_until TIMESTAMP WITH TIME ZONE;

-- Add hidden_forever field for permanent hiding
ALTER TABLE words ADD COLUMN IF NOT EXISTS hidden_forever BOOLEAN DEFAULT FALSE;

-- Add index for efficient queries filtering hidden words
CREATE INDEX IF NOT EXISTS idx_words_hidden ON words(user_id, hidden_forever, hidden_until) 
WHERE hidden_forever = TRUE OR hidden_until IS NOT NULL;

-- Comment for future reference
COMMENT ON COLUMN words.hidden_until IS 'Timestamp until which word is hidden from random pair (NULL = not hidden)';
COMMENT ON COLUMN words.hidden_forever IS 'If TRUE, word is permanently hidden from random pair';
