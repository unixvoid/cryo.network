GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0

cryon:
	$(GOC) cryon.go

run:
	go run cryon.go

stat:
	$(CGOR) $(GOC) $(GOFLAGS) cryon.go
