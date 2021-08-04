package main

import "regexp"

var responseMatchers = map[string]*regexp.Regexp{}

func init() {
	responseRegexps := map[string]string{
		"addsticker":       "Choose the sticker pack you're interested in\\.",
		"stickerPack_sent": "Alright! Now send me the sticker\\.",
		"image_sent":       "Thanks! Now send me an emoji that corresponds to your first sticker\\.",
		"emoji_sent":       "There we go\\. I've added your sticker to the pack",
		"done":             "OK, well done!",
	}

	for name, expression := range responseRegexps {
		responseMatchers[name] = regexp.MustCompile(expression)
	}
}
