-- Create users table
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    authorized BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create words table
CREATE TABLE IF NOT EXISTS words (
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    word TEXT NOT NULL,
    translation TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

-- Create basic indexes
CREATE INDEX IF NOT EXISTS idx_words_user_id ON words(user_id);
CREATE INDEX IF NOT EXISTS idx_words_created_at ON words(created_at);

