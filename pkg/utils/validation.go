package utils

import "strings"

func RemoveWhitespace(s string) string {
	return strings.ReplaceAll(s, " ", "")
}

func SplitEmail(s string) (string, string) {
	parts := strings.Split(s, "@")
	return parts[0], parts[1]
}

func IsEmail(s string) bool {
	return strings.Contains(s, "@")
}

// IsValidMajor checks if the major is valid
func IsValidMajor(major string) bool {
	return major == "information system" || major == "informatics"
}

// ValidateMajor validates the major field and returns an error message if invalid
func ValidateMajor(major string) (bool, string) {
	if major == "" {
		return false, "major is required"
	}
	if !IsValidMajor(major) {
		return false, "invalid major: must be 'information system' or 'informatics'"
	}
	return true, ""
}
