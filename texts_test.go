package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var responseTexts = map[string]string{
	"addsticker": `Choose the sticker pack you're interested in.`,
	"stickerPack_sent": `Alright! Now send me the sticker. The image file should be in PNG or WEBP format with a transparent layer and must fit into a 512x512 square (one of the sides must be 512px and the other 512px or less).

The sticker should use white stroke and shadow, exactly like in this .PSD example: StickerExample.psd (https://telegram.org/img/StickerExample.psd).

I recommend using Telegram for Web/Desktop when uploading stickers.`,
	"image_sent": `Thanks! Now send me an emoji that corresponds to your first sticker.

You can list several emoji in one message, but I recommend using no more than two per sticker.`,
	"emoji_sent": `There we go. I've added your sticker to the pack, it will become available to all Telegram users within about an hour. 

To add another sticker, send me the next sticker.
When you're done, simply send the /done command.`,
	"done": "OK, well done!",
}

func TestMatcher(t *testing.T) {
	for matcherName, matcher := range responseMatchers {
		t.Run(matcherName, func(t *testing.T) {
			a := require.New(t)

			for textName, text := range responseTexts {
				matches := matcher.MatchString(text)

				if textName == matcherName {
					a.True(matches)
				} else {
					a.False(matches)
				}
			}
		})
	}
}
