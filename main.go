package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"slices"
	"sort"
	"time"

	"github.com/huml-lang/go-huml"
	"github.com/mmcdole/gofeed"
)

const LASTREADHUML = "$HOME/.local/share/lastread.huml"
const CONFIGHUML = "$HOME/.config/rssreader.huml"

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

func getLastReadHUMLPath() string {
	return os.ExpandEnv(LASTREADHUML)
}

func getConfigHUMLPath() string {
	return os.ExpandEnv(CONFIGHUML)
}

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

	end := len(feed.Items)
	for i, v := range feed.Items {
		if v.PublishedParsed.Unix() <= lastRead[link] {
			end = i
			break
		}
	}

	return feed.Items[0:end], nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	bc, err := os.ReadFile(getConfigHUMLPath())
	if err != nil {
		log.Fatal(err)
	}
	var config Config
	if err := huml.Unmarshal(bc, &config); err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(getLastReadHUMLPath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			lastRead = make(map[string]int64)
			if err := saveLastRead(); err != nil {
				log.Fatal(err)
			}
		}
	}
	blr, err := os.ReadFile(getLastReadHUMLPath())
	if err != nil {
		log.Fatal(err)
	}
	if err := huml.Unmarshal(blr, &lastRead); err != nil {
		log.Fatal(err)
	}
	log.Println(config)
	if slices.Contains(os.Args, "bg") {
		for _, v := range config.Rss {
			if v.Important == Important {
				items, err := readFeed(v.Link)
				if err != nil {
					log.Println(err)
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
			fmt.Println(err)
			continue
		}
		for _, item := range items {
			fmt.Println(item.Title)
			fmt.Println(item.Link)
			scanner.Scan()
			if err := markRead(v.Link, *item.PublishedParsed); err != nil {
				log.Fatal(err)
			}
		}
	}
}
