CREATE TABLE IF NOT EXISTS candidates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    vision TEXT,
    mission TEXT,
    class VARCHAR(50),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    major VARCHAR(255) NOT NULL CHECK (major IN ('information system', 'informatics')),
    profile_picture VARCHAR(255),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_candidates_major ON candidates(major);
CREATE INDEX idx_candidates_user_id ON candidates(user_id);
