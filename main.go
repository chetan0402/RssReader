package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"sort"

	"github.com/fatih/color"
	"github.com/huml-lang/go-huml"
	"github.com/mmcdole/gofeed"
)

const (
	Important = iota
	Normal
	Low
)

type Config struct {
	Rss []struct {
		Link      string
		Important int
	}
}

var fp = gofeed.NewParser()

func readFeed(lr *lastRead, link string) ([]*gofeed.Item, error) {
	feed, err := fp.ParseURL(link)
	if err != nil {
		return nil, err
	}
	sort.Sort(feed)

	start := len(feed.Items)
	for i, v := range feed.Items {
		if lr.isRead(link, v.PublishedParsed) {
			start = i
			break
		}
		slog.Debug("skip", "title", feed.Title)
	}

	return feed.Items[start:len(feed.Items)], nil
}

func main() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if slices.Contains(os.Args, "-v") {
		opts.Level = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, opts)))
	scanner := bufio.NewScanner(os.Stdin)
	bc, err := os.ReadFile(getConfigHUMLPath())
	if err != nil {
		slog.Error("err", "err", err)
		os.Exit(1)
	}
	var config Config
	if err := huml.Unmarshal(bc, &config); err != nil {
		slog.Error("err", "err", err)
		os.Exit(1)
	}
	lr := newLastRead()
	slog.Debug("load", "config", config)
	if slices.Contains(os.Args, "bg") {
		for _, v := range config.Rss {
			if v.Important == Important {
				_, err := readFeed(lr, v.Link)
				if err != nil {
					slog.Error("fetch", "err", err)
					continue
				}
				if err := exec.Command("notify-send", "RSS", fmt.Sprintf("Articles from %v", v.Link)).Run(); err != nil {
					slog.Error("notify", "err", err)
				}
			}
		}
		os.Exit(0)
	}
	for _, v := range config.Rss {
		items, err := readFeed(lr, v.Link)
		if err != nil {
			slog.Error("err", "err", err)
			continue
		}
		for _, item := range items {
			fmt.Println(item.Title)
			color.Set(color.FgHiBlack)
			fmt.Println(item.Link)
			color.Unset()
			scanner.Scan()
			if err := lr.markRead(v.Link, *item.PublishedParsed); err != nil {
				slog.Error("err", "err", err)
			}
		}
	}
}
