package main

import (
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"slices"
	"sort"
	"time"

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

var lastRead map[string]int64
var fp = gofeed.NewParser()

func saveLastRead() error {
	blr, err := huml.Marshal(lastRead)
	if err != nil {
		return err
	}
	if err := os.WriteFile(getLastReadHUMLPath(), blr, 0644); err != nil {
		return err
	}
	return nil
}

func markRead(link string, t time.Time) error {
	if t.Unix() > lastRead[link] {
		lastRead[link] = t.Unix()
	}
	if err := saveLastRead(); err != nil {
		return err
	}
	return nil
}

func readFeed(link string) ([]*gofeed.Item, error) {
	feed, err := fp.ParseURL(link)
	if err != nil {
		return nil, err
	}
	sort.Sort(feed)

	start := len(feed.Items)
	for i, v := range feed.Items {
		if v.PublishedParsed.Unix() > lastRead[link] {
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
	if _, err := os.Stat(getLastReadHUMLPath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			lastRead = make(map[string]int64)
			if err := saveLastRead(); err != nil {
				slog.Error("savefailed", "err", err)
				os.Exit(1)
			}
		}
	}
	blr, err := os.ReadFile(getLastReadHUMLPath())
	if err != nil {
		slog.Error("err", "err", err)
		os.Exit(1)
	}
	if err := huml.Unmarshal(blr, &lastRead); err != nil {
		slog.Error("err", "err", err)
		os.Exit(1)
	}
	slog.Debug("load", "config", config)
	if slices.Contains(os.Args, "bg") {
		for _, v := range config.Rss {
			if v.Important == Important {
				items, err := readFeed(v.Link)
				if err != nil {
					slog.Error("fetch", "err", err)
					continue
				}
				exec.Command("notify-send", "RSS", fmt.Sprintf("%v articles from %v", len(items), v.Link))
			}
		}
		os.Exit(0)
	}
	for _, v := range config.Rss {
		items, err := readFeed(v.Link)
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
			if err := markRead(v.Link, *item.PublishedParsed); err != nil {
				slog.Error("err", "err", err)
			}
		}
	}
}
