package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"

	lt "github.com/jweslley/localtunnel"
)

var (
	errPortRequired = errors.New("Missing required argument: port")

	host      = flag.String("h", "https://localtunnel.me", "Upstream server providing forwarding")
	local     = flag.String("l", "localhost", "Tunnel traffic to this host instead of localhost")
	subdomain = flag.String("s", "", "Request this subdomain")
	port      = flag.Int("p", 0, "Internal http server port")
)

func fail(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: lt -p <PORT> [OPTION]...\n")
	fmt.Fprintf(os.Stderr, "localtunnel exposes your localhost to the world for easy testing and sharing!\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *port == 0 {
		usage()
		fail(errPortRequired)
	}

	c := lt.NewClient(*host)
	t := c.NewTunnel(*local, *port)

	if *subdomain == "" {
		fail(t.Open())
	} else {
		fail(t.OpenAs(*subdomain))
	}

	fmt.Printf("your url is: %s\n", t.URL())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		for s := range sig {
			fmt.Printf("%v received\n", s)
			t.Close()
		}
	}()

	<-t.Closing()
	fmt.Println("Bye! tunnel closed")
}
