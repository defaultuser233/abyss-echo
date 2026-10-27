APP      := abyss-echo
VERSION  ?= 1.0.0
LDFLAGS  := -s -w -X main.version=$(VERSION)

.PHONY: build run test dist clean

build:
	@mkdir -p bin
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(APP) .

run:
	go run .

test:
	go test ./...

dist:
	VERSION=$(VERSION) ./scripts/build-all.sh

clean:
	rm -rf bin dist
