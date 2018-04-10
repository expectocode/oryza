# Oryza

A private file host featuring a Go REST API. In progress.

#### How do I upload stuff?

 - `curl -X POST -F mimetype='image/png' -F extrainfo="my desktop" -F uploadfile=@/tmp/screenshot.png -F token=$token http://up.unix.porn/api/upload`
 - Message @oryzabot on Telegram
 - (TODO) Use the page at up.unix.porn

#### Done:

 - Backend upload
 - Backend registration
 - Backend file serving
 - cURL for any of the above (REST!)
 - Telegram bot upload (Document/Image/GIF)
 - CLI upload/screenshot script
 - Configurable with config.sh

#### To do:
 - Backend deletion
 - Backend details page display on the long URL
 - Web frontend upload
 - Web frontend file listin
 - Web frontend, full stop.
 - Syntax highlighting for code files


#### I want an account!

Ask me nicely. Or run your own instance of it.

#### How do I run my own instance?

Clone the code, edit the config values, figure out some hosting, and you should be alright. Dependencies for the code can be retrieved with `go get` (although Go seems to be a bit confused over how this should actually be done).
