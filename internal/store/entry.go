package store

import "time"

// Kind classifies what a secret is. Purely informational.
const (
	KindToken    = "Token"
	KindAPIKey   = "API Key"
	KindPassword = "Password"
	KindOther    = "Other"
)

// Kinds lists all selectable kinds in display order.
var Kinds = []string{KindToken, KindAPIKey, KindPassword, KindOther}

// Entry is a single stored secret.
type Entry struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	Value     string    `json:"value"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Masked returns a display-safe version of the secret value.
func (e Entry) Masked() string {
	n := len(e.Value)
	switch {
	case n == 0:
		return ""
	case n >= 16:
		return e.Value[:4] + "••••••••" + e.Value[n-4:]
	case n >= 8:
		return e.Value[:2] + "••••••" + e.Value[n-2:]
	default:
		return "••••••••"
	}
}
