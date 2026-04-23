package main

import "os"

const CONFIGHUML = "$HOME/.config/rssreader.huml"

func getConfigHUMLPath() string {
	return os.ExpandEnv(CONFIGHUML)
}
