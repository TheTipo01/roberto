balcon -i -o -n Roberto | ffmpeg -i pipe:0 -f s16le -ar 48000 -ac 2 pipe:1 | dca > ./temp/%1.dca