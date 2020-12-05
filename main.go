package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	aggrstate "github.com/aggr/aggr-cli/state"
	"github.com/mmcdole/gofeed"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/xeonx/timeago"
)

func main() {
	if len(os.Args) != 2 {
		log.Panic().Msg("Usage: aggr @username")
	}
	username := strings.TrimPrefix(os.Args[1], "@")
	if len(username) == 0 {
		log.Panic().Msg("Empty username")
	}

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	state, _ := aggrstate.LoadState(username)
	if state == nil {
		state = aggrstate.NewState(username)
	}

	if err := refreshState(context.Background(), httpClient, state); err != nil {
		log.Panic().Err(err).Msg("Couldn't refresh state")
	}
	go func() {
		if err := state.Save(); err != nil {
			log.Warn().Err(err).Msg("Couldn't save state to disk")
		}
	}()

	app := newApp(state.Feed)
	if err := app.Run(); err != nil {
		log.Panic().Err(err).Msg("Failed to run the TUI")
	}
}

func refreshState(ctx context.Context, httpClient http.Client, state *aggrstate.State) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, state.URL, nil)
	if err != nil {
		return err
	}
	if etag := state.ETag; len(etag) > 0 {
		req.Header.Set("if-none-match", etag)
	}
	if lm := state.LastModified; len(lm) > 0 {
		req.Header.Set("if-modified-since", lm)
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotModified {
		return nil
	}
	if sc := res.StatusCode; sc != http.StatusOK {
		return fmt.Errorf("unexpected status received from %s %s, got %d", req.Method, req.URL.String(), sc)
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(res.Body)
	if err != nil {
		return err
	}

	if etag := res.Header.Get("etag"); len(etag) > 0 {
		state.ETag = strings.TrimSpace(etag)
	}
	if lm := res.Header.Get("last-modified"); len(lm) > 0 {
		state.LastModified = strings.TrimSpace(lm)
	}
	state.Feed = feed

	return nil
}

func newApp(feed *gofeed.Feed) *tview.Application {
	app := tview.NewApplication()

	list := tview.NewList()

	for i, item := range feed.Items {
		mainText := item.Title

		secondaryText := "by " + item.Author.Name
		if pp := item.PublishedParsed; pp != nil {
			secondaryText += " | " + timeago.English.Format(*pp)
		}

		shortcut := rune(int('a') + i)

		list.AddItem(mainText, secondaryText, shortcut, nil)
	}

	list.SetChangedFunc(func(i int, _ string, _ string, _ rune) {
		if err := browser.OpenURL(feed.Items[i].Link); err != nil {
			log.Panic().Err(err).Msg("Failed to open the link in a browser")
		}
		app.Stop()
	})

	return app.SetRoot(list, true).EnableMouse(true)
}
