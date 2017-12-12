all: syntax.html

syntax.html: syntax.md
	cat pandoc_head.md syntax.md | pandoc -s -t html5 -o syntax.html

.PHONY: all
