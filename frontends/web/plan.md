# Web frontend for oryza upload service

Should allow users to login with their Oryza token, don't bother with session cookie.
Store token in cookie? Are cookies encrypted with TLS? Should be.

Then, use token to determine user and show appropriate pages generated on request (upload listing, edit details pages)

Already in Go ecosystem for the server with mux, now need some templating.
