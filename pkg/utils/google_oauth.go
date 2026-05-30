package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GoogleIDTokenClaims is the verified subset of fields we care about from a
// Google-issued ID token. Field names match Google's tokeninfo JSON response.
type GoogleIDTokenClaims struct {
	Sub           string `json:"sub"`            // stable user id
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"` // tokeninfo returns string "true"/"false"
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	HostedDomain  string `json:"hd"`             // present for Workspace accounts
	Aud           string `json:"aud"`            // our GOOGLE_CLIENT_ID
	Iss           string `json:"iss"`            // accounts.google.com or https://accounts.google.com
	Exp           string `json:"exp"`            // unix seconds, string
}

// IsEmailVerified converts the string-y verified flag into a bool.
func (c *GoogleIDTokenClaims) IsEmailVerified() bool {
	return strings.EqualFold(c.EmailVerified, "true")
}

// googleTokenInfoURL is Google's documented endpoint for verifying an ID
// token. It checks the signature, the issuer, and the expiration server-side,
// which lets us avoid pulling in a JWKS library.
//
// Ref: https://developers.google.com/identity/sign-in/web/backend-auth
const googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// VerifyGoogleIDToken sends the raw id_token to Google's tokeninfo endpoint
// and returns the parsed claims if it's valid AND issued to our client ID.
// expectedClientID is the GOOGLE_CLIENT_ID configured on the server.
func VerifyGoogleIDToken(idToken, expectedClientID string) (*GoogleIDTokenClaims, error) {
	if idToken == "" {
		return nil, errors.New("empty id_token")
	}
	if expectedClientID == "" {
		return nil, errors.New("GOOGLE_CLIENT_ID is not configured")
	}

	client := &http.Client{Timeout: 8 * time.Second}
	q := url.Values{"id_token": {idToken}}
	resp, err := client.Get(googleTokenInfoURL + "?" + q.Encode())
	if err != nil {
		return nil, fmt.Errorf("tokeninfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokeninfo returned status %d: %s", resp.StatusCode, string(body))
	}

	var claims GoogleIDTokenClaims
	if err := json.Unmarshal(body, &claims); err != nil {
		return nil, fmt.Errorf("tokeninfo decode failed: %w", err)
	}

	// Audience must match our client ID — guards against tokens minted for a
	// different OAuth app being replayed against us.
	if claims.Aud != expectedClientID {
		return nil, fmt.Errorf("id_token aud %q does not match expected client", claims.Aud)
	}

	// Issuer must be Google.
	if claims.Iss != "accounts.google.com" && claims.Iss != "https://accounts.google.com" {
		return nil, fmt.Errorf("id_token issuer %q is not Google", claims.Iss)
	}

	if !claims.IsEmailVerified() {
		return nil, errors.New("Google reports this email is not verified")
	}

	return &claims, nil
}
