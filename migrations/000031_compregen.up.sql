-- 1. Whitelist
CREATE TABLE compregen_eligible_candidates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id VARCHAR(20) UNIQUE NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    campus_email VARCHAR(255) NOT NULL,
    major VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Invite links
CREATE TABLE compregen_invite_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'used', 'expired', 'revoked')),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    used_at TIMESTAMPTZ
);
CREATE INDEX idx_compregen_invite_links_token ON compregen_invite_links(token);

-- 3. Registrations
CREATE TABLE compregen_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invite_link_id UUID NOT NULL REFERENCES compregen_invite_links(id),
    cabinet_name VARCHAR(255) NOT NULL,
    consent_accepted BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(20) NOT NULL DEFAULT 'submitted' CHECK (status IN ('submitted', 'approved', 'rejected')),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Registration members
CREATE TABLE compregen_registration_members (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_id UUID NOT NULL REFERENCES compregen_registrations(id),
    role VARCHAR(10) NOT NULL CHECK (role IN ('cp', 'vcp1', 'vcp2')),
    full_name VARCHAR(255) NOT NULL,
    student_id VARCHAR(20) NOT NULL,
    major VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    photo_upload_id UUID,
    UNIQUE(registration_id, role)
);

-- 5. Photo uploads
CREATE TABLE compregen_photo_uploads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registration_member_id UUID,
    storage_key VARCHAR(500) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes INT NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 6. Verify attempts (rate limiting)
CREATE TABLE compregen_verify_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    invite_link_id UUID NOT NULL REFERENCES compregen_invite_links(id),
    student_id_attempted VARCHAR(20) NOT NULL,
    email_attempted VARCHAR(255) NOT NULL,
    success BOOLEAN NOT NULL DEFAULT FALSE,
    attempted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);