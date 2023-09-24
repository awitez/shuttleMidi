build:
	@go build -o ShuttleMidi.app/Contents/MacOS/shuttlemidi

run: build
	@./ShuttleMidi.app/Contents/MacOS/shuttlemidi
test:
	@go test -v ./...
