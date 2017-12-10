all: draft.html server simpleechoback

draft.html: draft.md
	pandoc -s -thtml5 -o draft.html draft.md

server: server.go
	go build server.go

simpleechoback: simpleechoback.go
	go build simpleechoback.go

.PHONY: all
