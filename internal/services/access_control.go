package services

import (
	"Backend/internal/models"
	"regexp"
)

// access_control.go — authoritative "who is allowed to add data" check.
//
// Mirrors `Frontend-2026/src/lib/permissions.ts`. If you change the rule
// in one place, change it in the other. The frontend hides UI for blocked
// users; this is the layer that actually rejects the request.
//
// Rule:
//   - Role Admin (1)      → always allowed.
//   - Role Computizen (2) → must look like a real Faculty of Computer
//                           Science student account: 12-digit student_id,
//                           prefix in csMajorPrefixes, profile_completed.
//   - Everything else (Guest 6, PUFA/PUMA execs, anyone else) → blocked.

// csMajorPrefixes are the Student ID prefixes that mark the holder as a
// Faculty of Computer Science student. Matches the values in
// internal/services/auth_services.go's promotion logic and the frontend's
// CS_MAJOR_PREFIXES.
var csMajorPrefixes = map[string]struct{}{
	"001": {}, // Informatics
	"012": {}, // Information System
	"013": {}, // Visual Communication Design
	"025": {}, // Interior Design
}

var studentIDPattern = regexp.MustCompile(`^[0-9]{12}$`)

// HasFacultyStudentID reports whether the user's StudentID is a 12-digit
// numeric value whose major prefix is one of the recognised CS majors.
func HasFacultyStudentID(user *models.User) bool {
	if user == nil || user.StudentID == "" {
		return false
	}
	if !studentIDPattern.MatchString(user.StudentID) {
		return false
	}
	_, ok := csMajorPrefixes[user.StudentID[:3]]
	return ok
}

// CanCreateContent decides whether the given user is allowed to perform a
// "create data" action on the public app (project submission, aspiration,
// etc.). Returns the boolean verdict and a short user-facing reason when
// the verdict is false. The reason is safe to surface in API responses.
func CanCreateContent(user *models.User) (bool, string) {
	if user == nil {
		return false, "You must be signed in with a Faculty of Computer Science student account to add data."
	}

	if user.RoleID == models.RoleAdmin {
		return true, ""
	}

	if user.RoleID == models.RoleGuest {
		return false, "Your account is registered as a Guest. Only Faculty of Computer Science students (Computizens) and administrators can add data — you can still browse everything in read-only mode."
	}

	if user.RoleID != models.RoleComputizen {
		return false, "Only Faculty of Computer Science students (Computizens) and administrators can add data."
	}

	// Computizen branch — must look like a real CS student account.
	if !user.ProfileCompleted {
		return false, "Finish your onboarding (Student ID + batch year) to unlock writing data."
	}
	if !HasFacultyStudentID(user) {
		return false, "Adding data is reserved for Faculty of Computer Science student accounts. Your Student ID isn't recognised as a CS major."
	}
	return true, ""
}
