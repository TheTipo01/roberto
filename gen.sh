#!/bin/bash
cd temp
WINEPREFIX=~/.PlayOnLinux/wineprefix/tts wine c:/balcon/balcon -w $1.wav -n Roberto -t "$2"
ffmpeg -i $1.wav -f s16le -ar 48000 -ac 2 pipe:1 | dca > $1.dca
rm $1.wav