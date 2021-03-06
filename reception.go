package main

import (
	"flag"
	"fmt"

	net_http "net/http"

	"runtime"

	miekg_dns "github.com/miekg/dns"
	"github.com/ninech/reception/common"
	"github.com/ninech/reception/dns"
	"github.com/ninech/reception/docker"
	"github.com/ninech/reception/http"
)

var config = &common.Config{
	Projects: common.NewProjects(),
}

var (
	Tag         string = "SNAPSHOT"
	BuildDate   string
	Commit      string
	Branch      string
	ShowVersion bool
)

func init() {
	flag.StringVar(
		&config.HTTPBindAddress,
		"http.address",
		"localhost:80",
		"Defines on which address and port the HTTP daemon listens.")
	flag.StringVar(
		&config.DNSBindAddress,
		"dns.address",
		"localhost:53",
		"Defines on which address and port the HTTP daemon listens.")
	flag.StringVar(
		&config.TLD,
		"tld",
		"docker.",
		"Defines on which TLD to react for HTTP and DNS requests. Should end with a \".\" .")
	flag.StringVar(
		&config.DockerEndpoint,
		"docker.endpoint",
		"unix:///var/run/docker.sock",
		"How reception talks to Docker.")
	BoolFlag(&ShowVersion, "version", false, "Show version.")
}

func BoolFlag(p *bool, name string, value bool, usage string) {
	flag.BoolVar(p, name, value, usage)
	flag.BoolVar(p, name[:1], value, usage)
}

func main() {
	fmt.Println("(c) 2017 Nine Internet Solutions AG")

	flag.Parse()

	if ShowVersion {
		showVersionInfo()
		return
	}

	go runHttpFrontend()
	go runDns()

	runDockerClient()
}

func runDns() {
	handler := dns.Handler{
		Config: config,
	}

	tld := config.TLD
	if "." != tld[len(tld)-1:] {
		tld += "."
	}

	miekg_dns.HandleFunc(tld, handler.ServeDns)

	addr := config.DNSBindAddress
	fmt.Printf("Listening on '%v' for DNS requests.\n", addr)

	srv := &miekg_dns.Server{Addr: addr, Net: "udp"}
	err := srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func runDockerClient() {
	client := docker.Client{
		Config: config,
	}
	err := client.Launch()
	if err != nil {
		panic(err)
	}
}

func runHttpFrontend() {
	frontend := &net_http.Server{
		Addr: config.HTTPBindAddress,
		Handler: http.BackendHandler{
			Config: config,
		},
	}

	fmt.Printf("Listening on '%v' for HTTP traffic.\n", frontend.Addr)

	err := frontend.ListenAndServe()
	if err != nil {
		panic(err)
	} else {
		defer frontend.Close()
	}
}

func showVersionInfo() {
	fmt.Println("Version:    ", Tag)
	fmt.Println("Build Date: ", BuildDate)
	fmt.Println("Commit:     ", Commit)
	fmt.Println("Branch:     ", Branch)
	fmt.Println("Go Version: ", runtime.Version())
}
