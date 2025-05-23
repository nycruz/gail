run: # Run the project
	go run *.go

build: # Build the project
	go build -o ./bin/gail *.go

dev-gpt: # Run ChatGPT standard model
	go run *.go --model=gpt

dev-gpt-o: # Run ChatGPT o-series model
	go run *.go --model=gpt-o

dev-claude: # Run Claude standard model
	go run *.go --model=claude

.PHONY: run build dev-gpt dev-gpt-o dev-claude
