// Package domain contains core business entities.
// This layer has NO external dependencies - it's the innermost layer of the architecture.
package domain

// Identity represents a user identity.
type Identity struct {
	ID          string
	DisplayName string
	UniqueName  string
	URL         string
	ImageURL    string
}

// ShortName returns a display-friendly short name.
func (i Identity) ShortName() string {
	if i.DisplayName != "" {
		return i.DisplayName
	}
	return i.UniqueName
}
