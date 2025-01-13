run:
	go run *.go

build:
	go build -o ./bin/gail *.go

dev-gpt:
	go run *.go --model=gpt

dev-claude:
	go run *.go --model=claude

# dev-gpt-o1:
# 	go run *.go --model=gpt-o1
#
# dev-gpt-o1-mini:
# 	go run *.go --model=gpt-o1-mini


.PHONY: all test clean run dev-gpt dev-claude dev-gpto1 dev-gpto1-mini
