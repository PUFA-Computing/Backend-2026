CREATE TABLE IF NOT EXISTS votes (
    id SERIAL PRIMARY KEY,
    voter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    candidate_id INT NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT unique_vote_per_major UNIQUE (voter_id)
);

CREATE OR REPLACE FUNCTION check_same_major()
RETURNS TRIGGER AS $$
BEGIN
    IF (
        (SELECT major FROM users WHERE id = NEW.voter_id) <>
        (SELECT major FROM candidates WHERE id = NEW.candidate_id)
    ) THEN
        RAISE EXCEPTION 'Voter and candidate must have the same major';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_same_major
BEFORE INSERT ON votes
FOR EACH ROW
EXECUTE FUNCTION check_same_major();

CREATE INDEX idx_votes_voter_id ON votes(voter_id);
CREATE INDEX idx_votes_candidate_id ON votes(candidate_id);
