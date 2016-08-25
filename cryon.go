package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/miekg/dns"
	"github.com/unixvoid/glogger"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	Cryo struct {
		Loglevel       string
		Port           int
		DomainName     string
		DefaultAddress string
		DefaultCname   string
		DefaultAaaa    string
		DefaultTTL     uint32
	}
}

var (
	config = Config{}
)

func main() {
	readConf()
	initLogger(config.Cryo.Loglevel)
	// format the string to be :port
	fPort := fmt.Sprint(":", config.Cryo.Port)

	udpServer := &dns.Server{Addr: fPort, Net: "udp"}
	tcpServer := &dns.Server{Addr: fPort, Net: "tcp"}
	glogger.Info.Println("started server on", config.Cryo.Port)
	dns.HandleFunc(".", func(w dns.ResponseWriter, req *dns.Msg) {

		switch req.Question[0].Qtype {
		case 1:
			glogger.Debug.Println("'A' request recieved, continuing")
			route(w, req)
		case 5:
			glogger.Debug.Println("Routing 'CNAME' request")
			go cnameresolve(w, req)
			break
		case 28:
			glogger.Debug.Println("Routing 'AAAA' request")
			//go aaaaresolve(w, req)
			break
		default:
			glogger.Debug.Println("Not 'A' request")
			break
		}

	})

	go func() {
		glogger.Error.Println(udpServer.ListenAndServe())
	}()
	glogger.Error.Println(tcpServer.ListenAndServe())
}

func readConf() {
	// init config file
	err := gcfg.ReadFileInto(&config, "config.gcfg")
	if err != nil {
		panic(fmt.Sprintf("Could not load config.gcfg, error: %s\n", err))
	}
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
	glogger.Cluster.Println(hostname)
	domain := parseHostname(hostname)

	// craft response
	rr := new(dns.A)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: config.Cryo.DefaultTTL}
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

func cnameresolve(w dns.ResponseWriter, req *dns.Msg) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)
	//domain := parseHostname(hostname)

	// craft response
	rr := new(dns.CNAME)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeCNAME, Class: dns.ClassINET, Ttl: config.Cryo.DefaultTTL}
	// return default cname
	rr.Target = config.Cryo.DefaultCname

	// craft reply
	rep := new(dns.Msg)
	rep.SetReply(req)
	rep.Answer = append(rep.Answer, rr)

	// send it
	w.WriteMsg(rep)
	return
}

func aaaaresolve(w dns.ResponseWriter, req *dns.Msg) {
	hostname := req.Question[0].Name
	glogger.Cluster.Println(hostname)
	//domain := parseHostname(hostname)

	// craft response
	rr := new(dns.AAAA)
	rr.Hdr = dns.RR_Header{Name: hostname, Rrtype: dns.TypeAAAA, Class: dns.ClassINET, Ttl: config.Cryo.DefaultTTL}
	//addr := strings.TrimSuffix(lookup, "\n")
	addr := strings.TrimSuffix(config.Cryo.DefaultAaaa, "\n")
	rr.AAAA = net.ParseIP(addr)
	// return default cname

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
	if strings.Contains(hostname, config.Cryo.DomainName) {
		parsedHostname := strings.Replace(hostname, fmt.Sprintf("%s.", config.Cryo.DomainName), "", -1)
		glogger.Debug.Printf("parsedHostname: %s\n", parsedHostname)

		// count the number of '.'s so we can only grab the ip from the end
		//   this ip will end up being the last 4 '.'s

		// make sure we have proper syntax. If less than 4 '.'s remaining, there is no ip
		dotCount := strings.Count(parsedHostname, ".")
		if dotCount < 4 {
			glogger.Debug.Println("bad syntax, returning early")
			return config.Cryo.DefaultAddress
		}
		s := strings.Split(parsedHostname, ".")
		ip := fmt.Sprintf("%s.%s.%s.%s", s[dotCount-4], s[dotCount-3], s[dotCount-2], s[dotCount-1])
		glogger.Debug.Printf("Parsed ip: %s\n", ip)
		return ip
	} else {
		// the domain was not in the query, return home
		return config.Cryo.DefaultAddress
	}
}
