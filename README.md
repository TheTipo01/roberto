# roberto
[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/roberto)](https://goreportcard.com/report/github.com/TheTipo01/roberto)
[![Build Status](https://travis-ci.com/TheTipo01/roberto.svg?branch=master)](https://travis-ci.com/TheTipo01/roberto)

Discord TTS bot

Dependencies: [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [ffmpeg](https://ffmpeg.org/download.html) and Loquendo Roberto SAPI voice

For download, see releases.

## Bot commands

`.say <text>` - Says text out loud

`.bestemmia <n>` - Generates a bestemmia n times

`.treno <train number>` - Fakes train announcement given it's number

`.covid` - Says covid data out loud for current day in Italy

`.preghiera` - Randomly select a custom command

`.custom <custom command> <text>` - Creates a custom command to say text out loud. The bot will replace <god> with a random god and <dict> with a random adjective

`.rmcustom <custom command>` - Removes a custom command