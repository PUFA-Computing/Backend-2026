package app

import (
	"Backend/internal/database"
	"Backend/internal/models"
	"context"
	"errors"
	"time"
)

// ── Invite Links ───────────────────────────────────────

func GetInviteLinkByToken(token string) (*models.CompregenInviteLink, error) {
	var link models.CompregenInviteLink
	query := `
		SELECT id, token, status, created_by, created_at, expires_at, used_at
		FROM compregen_invite_links
		WHERE token = $1`
	err := database.DB.QueryRow(context.Background(), query, token).Scan(
		&link.ID, &link.Token, &link.Status, &link.CreatedBy,
		&link.CreatedAt, &link.ExpiresAt, &link.UsedAt,
	)
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func CreateInviteLink(token, createdBy string) (*models.CompregenInviteLink, error) {
	var link models.CompregenInviteLink
	query := `
		INSERT INTO compregen_invite_links (token, created_by)
		VALUES ($1, $2)
		RETURNING id, token, status, created_by, created_at, expires_at, used_at`
	err := database.DB.QueryRow(context.Background(), query, token, createdBy).Scan(
		&link.ID, &link.Token, &link.Status, &link.CreatedBy,
		&link.CreatedAt, &link.ExpiresAt, &link.UsedAt,
	)
	return &link, err
}

func MarkInviteLinkUsed(linkID string) error {
	query := `
		UPDATE compregen_invite_links
		SET status = 'used', used_at = NOW()
		WHERE id = $1`
	_, err := database.DB.Exec(context.Background(), query, linkID)
	return err
}

// ── Verify / Rate Limit ────────────────────────────────

func CheckEligibleCandidateByStudentID(studentID string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM compregen_eligible_candidates WHERE student_id = $1)`
    err := database.DB.QueryRow(context.Background(), query, studentID).Scan(&exists)
    return exists, err
}

func RecordVerifyAttempt(inviteLinkID, studentID, email string, success bool) error {
	query := `
		INSERT INTO compregen_verify_attempts (invite_link_id, student_id_attempted, email_attempted, success)
		VALUES ($1, $2, $3, $4)`
	_, err := database.DB.Exec(context.Background(), query, inviteLinkID, studentID, email, success)
	return err
}

// CheckRateLimit returns (isLocked, attemptsRemaining, error)
// Locked = 5 failed attempts within last 30 minutes
func CheckRateLimit(inviteLinkID string) (bool, int, error) {
	cooldown := 30 * time.Minute
	windowStart := time.Now().Add(-cooldown)

	var failedCount int
	query := `
		SELECT COUNT(*) FROM compregen_verify_attempts
		WHERE invite_link_id = $1
		  AND success = FALSE
		  AND attempted_at > $2`
	err := database.DB.QueryRow(context.Background(), query, inviteLinkID, windowStart).Scan(&failedCount)
	if err != nil {
		return false, 0, err
	}

	maxAttempts := 5
	if failedCount >= maxAttempts {
		return true, 0, nil
	}
	return false, maxAttempts - failedCount, nil
}

// ── Photo Uploads ──────────────────────────────────────

func SavePhotoUpload(storageKey, mimeType string, sizeBytes int) (*models.CompregenPhotoUpload, error) {
	var upload models.CompregenPhotoUpload
	query := `
		INSERT INTO compregen_photo_uploads (storage_key, mime_type, size_bytes)
		VALUES ($1, $2, $3)
		RETURNING id, registration_member_id, storage_key, mime_type, size_bytes, uploaded_at`
	err := database.DB.QueryRow(context.Background(), query, storageKey, mimeType, sizeBytes).Scan(
		&upload.ID, &upload.RegistrationMemberID, &upload.StorageKey,
		&upload.MimeType, &upload.SizeBytes, &upload.UploadedAt,
	)
	return &upload, err
}

func GetPhotoUploadByID(id string) (*models.CompregenPhotoUpload, error) {
	var upload models.CompregenPhotoUpload
	query := `
		SELECT id, registration_member_id, storage_key, mime_type, size_bytes, uploaded_at
		FROM compregen_photo_uploads WHERE id = $1`
	err := database.DB.QueryRow(context.Background(), query, id).Scan(
		&upload.ID, &upload.RegistrationMemberID, &upload.StorageKey,
		&upload.MimeType, &upload.SizeBytes, &upload.UploadedAt,
	)
	return &upload, err
}

// ── Registration ───────────────────────────────────────

func CreateRegistration(inviteLinkID, cabinetName string, members map[string]models.CompregenMemberInput) (string, error) {
	ctx := context.Background()

	tx, err := database.DB.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var registrationID string
	err = tx.QueryRow(ctx, `
		INSERT INTO compregen_registrations (invite_link_id, cabinet_name, consent_accepted)
		VALUES ($1, $2, TRUE)
		RETURNING id`,
		inviteLinkID, cabinetName,
	).Scan(&registrationID)
	if err != nil {
		return "", err
	}

	for role, m := range members {
		ok, err := CheckEligibleCandidateByStudentID(m.StudentID)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", errors.New("not_whitelisted:" + role)
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO compregen_registration_members
				(registration_id, role, full_name, student_id, major, phone_number, nationality, photo_upload_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			registrationID, role,
			m.FullName, m.StudentID, m.Major,
			m.PhoneNumber, m.Nationality, m.PhotoUploadID,
		)
		if err != nil {
			return "", err
		}
	}

	return registrationID, tx.Commit(ctx)
}

func GetAllRegistrations() ([]*models.CompregenRegistrationResponse, error) {
	ctx := context.Background()

	rows, err := database.DB.Query(ctx, `
		SELECT r.id, r.cabinet_name, r.submitted_at,
		       m.role, m.full_name, m.student_id, m.major, m.phone_number, m.nationality, m.photo_upload_id
		FROM compregen_registrations r
		JOIN compregen_registration_members m ON m.registration_id = r.id
		ORDER BY r.submitted_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	regMap := make(map[string]*models.CompregenRegistrationResponse)
	var order []string

	for rows.Next() {
		var (
			id, cabinetName, role, fullName, studentID, major, nationality, phone string
			submittedAt                                               time.Time
			photoUploadID                                             *string
		)
		if err := rows.Scan(&id, &cabinetName, &submittedAt,
			&role, &fullName, &studentID, &major, &phone, &nationality, &photoUploadID); err != nil {
			return nil, err
		}

		if _, exists := regMap[id]; !exists {
			regMap[id] = &models.CompregenRegistrationResponse{
				ID:          id,
				CabinetName: cabinetName,
				SubmittedAt: submittedAt,
				Members:     make(map[string]models.CompregenMemberResponse),
			}
			order = append(order, id)
		}

		regMap[id].Members[role] = models.CompregenMemberResponse{
			FullName:      fullName,
			StudentID:     studentID,
			Major:         major,
			PhoneNumber:   phone,
			Nationality:   nationality,
			PhotoUploadID: photoUploadID,
		}
	}

	result := make([]*models.CompregenRegistrationResponse, 0, len(order))
	for _, id := range order {
		result = append(result, regMap[id])
	}
	return result, nil
}

func RegistrationExistsForLink(inviteLinkID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM compregen_registrations WHERE invite_link_id = $1)`
	err := database.DB.QueryRow(context.Background(), query, inviteLinkID).Scan(&exists)
	return exists, err
}
