package main

import "os"

const LASTREADHUML = "$HOME/.local/share/lastread.huml"
const CONFIGHUML = "$HOME/.config/rssreader.huml"

func getLastReadHUMLPath() string {
	return os.ExpandEnv(LASTREADHUML)
}

func getConfigHUMLPath() string {
	return os.ExpandEnv(CONFIGHUML)
}
