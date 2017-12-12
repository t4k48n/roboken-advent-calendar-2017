all: draft.html

draft.html: draft.md
	cat pandoc_head.md draft.md | pandoc -s -t html5 -o draft.html

.PHONY: all
