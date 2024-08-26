run:
	go run *.go

build:
	go build -o ./bin/gail *.go

dev-gpt:
	go run *.go --model=gpt

dev-claude:
	go run *.go --model=claude

.PHONY: all test clean run dev-gpt dev-claude
