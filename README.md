# roberto
[![Go Report Card](https://goreportcard.com/badge/github.com/TheTipo01/roberto)](https://goreportcard.com/report/github.com/TheTipo01/roberto)

Discord TTS bot

[I have an instance hosted locally, if you want to try the bot out!](https://discord.com/api/oauth2/authorize?client_id=587761918865834023&permissions=3145728&scope=bot%20applications.commands)

## Notes
- We now use slash commands (from release [0.6.0](https://github.com/TheTipo01/roberto/releases/tag/0.6.0))
- Dependencies: [DCA](https://github.com/bwmarrin/dca/tree/master/cmd/dca), [ffmpeg](https://ffmpeg.org/download.html) and Loquendo Roberto SAPI voice

To download, see releases page.

## Bot commands

`/say <text>` - Says text out loud

`/bestemmia <n>` - Generates a bestemmia n times

`/treno <train number>` - Fakes train announcement given its number

`/covid` - Says covid data out loud for current day in Italy

`/preghiera` - Randomly select a custom command

`/addcustom <custom command> <text>` - Creates a custom command to say inputted text out loud. The bot will replace <god> with a random evangelical figure and <dict> with a random adjective

`/custom <custom_command>` - Executes a previously created custom command

`/rmcustom <custom command>` - Removes a custom command
