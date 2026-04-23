package main

import (
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/huml-lang/go-huml"
)

const LASTREADHUML = "$HOME/.local/share/lastread.huml"

func getLastReadHUMLPath() string {
	return os.ExpandEnv(LASTREADHUML)
}

type lastRead struct {
	mp map[string]int64
}

func newLastRead() *lastRead {
	l := &lastRead{}
	if _, err := os.Stat(getLastReadHUMLPath()); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			l.mp = make(map[string]int64)
			if err := l.saveLastRead(); err != nil {
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
	if err := huml.Unmarshal(blr, &l.mp); err != nil {
		slog.Error("err", "err", err)
		os.Exit(1)
	}
	return l
}

func (l *lastRead) saveLastRead() error {
	blr, err := huml.Marshal(l.mp)
	if err != nil {
		return err
	}
	if err := os.WriteFile(getLastReadHUMLPath(), blr, 0644); err != nil {
		return err
	}
	return nil
}

func (l *lastRead) markRead(link string, t time.Time) error {
	if t.Unix() > l.mp[link] {
		l.mp[link] = t.Unix()
	}
	if err := l.saveLastRead(); err != nil {
		return err
	}
	return nil
}

func (l *lastRead) isRead(link string, t *time.Time) bool {
	return t.Unix() < l.mp[link]
}
