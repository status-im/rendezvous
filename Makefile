deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

test:
	go test ./...

image:
	docker build . -t statusteam/rendezvous:latest

push:
	docker push statusteam/rendezvous:latest
