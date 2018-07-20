#!/usr/bin/env bash
source ~/.config/oryza
# This should define $token

# This script requires `jq`, since the server replies with a JSON response

screenshooter() {
	# Feel free to switch this out for a different screenshot tool and set $ to the saved image
	# I would have used something like that, but couldn't get those tools to
	# actually work when this script was bound to a key or called from rofi
	xfce4-screenshooter
	file=$(find /tmp -cmin -1 -iname "Screenshot*.png" 2>/dev/null|head -n 1) # created within 1 min, name screenshot
	mime="image/png"
	if [[ -z $file ]]; then # if it is empty
		# cancel
		notify-send "cancelled screenshot upload, no matching file"
		exit 1
	fi
}

uploadfile() {
	if [[ -f "$location" ]]; then
		echo "doing file2"
		file="$location"
		echo "doing file3"
		mime=$(file --mime-type --brief $file)
	else
		echo "File not found."
		exit 1
	fi

}

if [[ "$#" -eq 1 ]]; then
	# we have one argument. it's a file.
	echo "doing file $file"
	location="$@"
	uploadfile
	echo "doing file $file with mime $mime"
else
	screenshooter
fi

# Upload and notify
resp=$(curl --silent -X POST -F mimetype="$mime" -F uploadfile="@$file" -F token=$token https://up.unix.porn/api/upload)
# Where would we source extra info from? CLI dialog? args?
if [[ -z $resp ]]; then # if it is empty
	# cancel
	notify-send "failed to upload, empty response"
	exit 1
fi
notify-send "$resp"
url=$(echo "$resp" | jq -r .url)
xdg-open "$url"
xclip <<< "$url"
xclip -selection clipboard <<< "$url"
