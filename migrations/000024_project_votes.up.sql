CREATE TABLE IF NOT EXISTS project_votes (
    id SERIAL PRIMARY KEY,
    project_id INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_user_vote UNIQUE (user_id, project_id)
);

CREATE INDEX IF NOT EXISTS idx_project_votes_project_id ON project_votes(project_id);
CREATE INDEX IF NOT EXISTS idx_project_votes_user_id ON project_votes(user_id);
CREATE INDEX IF NOT EXISTS idx_project_votes_created_at ON project_votes(created_at DESC);
