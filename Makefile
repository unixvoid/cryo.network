GOC=go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0
DOCKER_PREFIX="sudo"
IMAGE_NAME="unixvoid/cryon"
GIT_HASH=$(shell git rev-parse HEAD | head -c 10)

cryon:
	$(GOC) cryon.go

dependencies:
	go get github.com/miekg/dns
	go get github.com/unixvoid/glogger
	go get gopkg.in/gcfg.v1

run:
	go run cryon.go

docker:
	$(MAKE) stat
	mkdir -p stage.tmp/
	cp bin/cryon* stage.tmp/cryon
	cp Dockerfile stage.tmp/
	cd stage.tmp/ && \
		$(DOCKER_PREFIX) docker build -t $(IMAGE_NAME) .

aci:
	$(MAKE) stat
	mkdir -p stage.tmp/cryon-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/cryon-layout/rootfs/
	cp bin/cryon* stage.tmp/cryon-layout/rootfs/cryon
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/cryon-layout/rootfs/
	sed -i "s/\$$DIFF/$(GIT_HASH)/g" stage.tmp/cryon-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/cryon-layout/rootfs/
	cp deps/manifest.json stage.tmp/cryon-layout/manifest
	cd stage.tmp/ && \
		actool build cryon-layout cryon.aci && \
		mv cryon.aci ../
	@echo "cryon.aci built"

travisaci:
	wget https://github.com/appc/spec/releases/download/v0.8.7/appc-v0.8.7.tar.gz
	tar -zxf appc-v0.8.7.tar.gz
	$(MAKE) stat
	mkdir -p stage.tmp/cryon-layout/rootfs/
	tar -zxf deps/rootfs.tar.gz -C stage.tmp/cryon-layout/rootfs/
	cp bin/cryon* stage.tmp/cryon-layout/rootfs/cryon
	chmod +x deps/run.sh
	cp deps/run.sh stage.tmp/cryon-layout/rootfs/
	sed -i "s/\$$DIFF/$(GIT_HASH)/g" stage.tmp/cryon-layout/rootfs/run.sh
	cp config.gcfg stage.tmp/cryon-layout/rootfs/
	cp deps/manifest.json stage.tmp/cryon-layout/manifest
	cd stage.tmp/ && \
		../appc-v0.8.7/actool build cryon-layout cryon.aci && \
		mv cryon.aci ../
	@echo "cryon.aci built"

testaci:
	deps/testrkt.sh

stat:
	mkdir -p bin/
	$(CGOR) $(GOC) $(GOFLAGS) -o bin/cryon-$(GIT_HASH)-linux-amd64 cryon.go

clean:
	rm -rf bin/
	rm -f cryon
	rm -f cryon.aci
