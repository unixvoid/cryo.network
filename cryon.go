package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"github.com/unixvoid/glogger"
)

func main() {
	initLogger("debug")
	// format the string to be :port
	port := 53
	fPort := fmt.Sprint(":", port)

	udpServer := &dns.Server{Addr: fPort, Net: "udp"}
	tcpServer := &dns.Server{Addr: fPort, Net: "tcp"}
	glogger.Info.Println("started server on", port)
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {
		route(w, req)
	})

	go func() {
		glogger.Error.Println(udpServer.ListenAndServe())
	}()
	glogger.Error.Println(tcpServer.ListenAndServe())
}

func initLogger(logLevel string) {
	// init logger
	if logLevel == "debug" {
		glogger.LogInit(os.Stdout, os.Stdout, os.Stdout, os.Stderr)
	} else if logLevel == "cluster" {
		glogger.LogInit(os.Stdout, os.Stdout, ioutil.Discard, os.Stderr)
	} else if logLevel == "info" {
		glogger.LogInit(os.Stdout, ioutil.Discard, ioutil.Discard, os.Stderr)
	} else {
		glogger.LogInit(ioutil.Discard, ioutil.Discard, ioutil.Discard, os.Stderr)
	}
}

func route(w dns.ResponseWriter, req *dns.Msg) {
	// async run proxy task
	go resolve(w, req)
}

func resolve(w dns.ResponseWriter, req *dns.Msg) {
	hostname := req.Question[0].Name
	glogger.Debug.Println(hostname)

	// craft response
	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	addr := strings.TrimSuffix("127.0.0.1", "\n")
	rr.A = net.ParseIP(addr)

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}
