GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX="sudo"
IMAGE_NAME="unixvoid/cryo.network"

cryon:
	$(GOC) cryon.go

run:
	go run cryon.go

docker:
	$(MAKE) stat
	$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

clean:
	rm -f cryon

stat:
	$(CGOR) $(GOC) $(GOFLAGS) cryon.go
