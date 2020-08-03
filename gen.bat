balcon -w temp\%1.wav -n Roberto -t %2
ffmpeg -i ./temp/%1.wav -f s16le -ar 48000 -ac 2 pipe:1 | dca > ./temp/%1.dca