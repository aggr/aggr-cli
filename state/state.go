package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mmcdole/gofeed"
)

type State struct {
	Username string `json:"username"`
	URL      string `json:"url"`
	// optional fields that might be filled up down the road
	ETag         string       `json:"etag,omitempty"`
	LastModified string       `json:"lastmodified,omitempty"`
	Feed         *gofeed.Feed `json:"feed,omitempty"`
}

func NewState(username string) *State {
	return &State{
		Username: username,
		URL:      fmt.Sprintf("https://aggr.md/@%s.json", username),
	}
}

func LoadState(username string) (*State, error) {
	f, err := os.OpenFile(getPathForUsername(username), os.O_RDONLY, 0755)
	if err != nil {
		return nil, err
	}

	var s State
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	return &s, nil
}

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
