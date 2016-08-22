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

var (
	domainName = "cryo.network"
)

func main() {
	initLogger("debug")
	// format the string to be :port
	port := 8053
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
	domain := parseHostname(hostname)

	// craft response
	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0}
	addr := strings.TrimSuffix(domain, "\n")
	rr.A = net.ParseIP(addr)

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}

func parseHostname(hostname string) string {
	// we expect at least domainName at the end. parse it out
	if strings.Contains(hostname, domainName) {
		parsedHostname := strings.Replace(hostname, fmt.Sprintf("%s.", domainName), "", -1)
		glogger.Debug.Printf("parsedHostname: %s\n", parsedHostname)

		// count the number of '.'s so we can only grab the ip from the end
		//   this ip will end up being the last 4 '.'s

		// make sure we have proper syntax. If less than 4 '.'s remaining, there is no ip
		dotCount := strings.Count(parsedHostname, ".")
		if dotCount < 4 {
			glogger.Debug.Println("bad syntax, returning early")
			return "127.0.0.1"
		}
		s := strings.Split(parsedHostname, ".")
		ip := fmt.Sprintf("%s.%s.%s.%s", s[dotCount-4], s[dotCount-3], s[dotCount-2], s[dotCount-1])
		glogger.Debug.Printf("Parsed ip: %s\n", ip)
		return ip
	} else {
		// the domain was not in the query, return home
		return "127.0.0.1"
	}
}
