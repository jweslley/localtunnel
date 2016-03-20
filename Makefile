PROGRAM=lt
VERSION=0.1.0
LDFLAGS="-X main.programVersion=$(VERSION)"

all: test

build:
	go build -o $(PROGRAM) ./cmd/...

test:
	go test -v ./...

qa:
	go vet
	golint
	go test -coverprofile=.cover~
	go tool cover -html=.cover~

dist:
	@for os in linux darwin; do \
		for arch in 386 amd64; do \
			target=$(PROGRAM)-$$os-$$arch-$(VERSION); \
			echo Building $$target; \
			GOOS=$$os GOARCH=$$arch go build -ldflags $(LDFLAGS) -o $$target/$(PROGRAM) ; \
			cp ./README.md ./LICENSE $$target; \
			tar -zcf $$target.tar.gz $$target; \
			rm -rf $$target;                   \
		done                                 \
	done

clean:
	rm -rf *.tar.gz
