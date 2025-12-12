-- Remove fields for hiding words

-- Drop index
DROP INDEX IF EXISTS idx_words_hidden;

-- Remove columns
ALTER TABLE words DROP COLUMN IF EXISTS hidden_forever;
ALTER TABLE words DROP COLUMN IF EXISTS hidden_until;
