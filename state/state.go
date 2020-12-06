package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmcdole/gofeed"
)

// State contains cache information that can be re-used later on. A state is
// scoped around a single username.
type State struct {
	Username string `json:"username"`
	URL      string `json:"url"`
	// optional fields that might be filled up down the road
	ETag         string       `json:"etag,omitempty"`
	LastModified string       `json:"lastmodified,omitempty"`
	Feed         *gofeed.Feed `json:"feed,omitempty"`
}

// NewState creates a new state based on a specific username.
func NewState(username string) *State {
	return &State{
		Username: username,
		URL:      fmt.Sprintf("https://aggr.md/@%s.json", username),
	}
}

// LoadState loads a state from disk. The path is computed from a username.
func LoadState(username string) (*State, error) {
	f, err := os.Open(getPathForUsername(username))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s State
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

// Save serializes the state and dumps it to the disk.
func (s *State) Save() error {
	f, err := os.Create(getPathForUsername(s.Username))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(s)
}

func getPathForUsername(username string) string {
	return filepath.Join(os.TempDir(), "aggr-cli-feed-"+username)
}
