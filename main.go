package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/pkg/browser"
	"github.com/rivo/tview"
	"github.com/xeonx/timeago"
)

func main() {
	if len(os.Args) != 2 {
		panic("Usage: aggr @username")
	}
	username := os.Args[1]

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	// Fetch feed

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("https://aggr.md/%s.json", username),
		nil,
	)
	if err != nil {
		panic(err)
	}
	res, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if sc := res.StatusCode; sc != http.StatusOK {
		panic(fmt.Errorf("Unexpected status received from %s %s, got %d", req.Method, req.URL.String(), sc))
	}

	fp := gofeed.NewParser()
	feed, err := fp.Parse(res.Body)
	if err != nil {
		panic(err)
	}

	// TUI

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
			panic(err)
		}
		app.Stop()
	})

	if err := app.SetRoot(list, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
