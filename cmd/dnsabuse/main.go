package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dnsproject/internal/services/dice"
	"dnsproject/internal/services/fx"
	"dnsproject/internal/services/random"
	"dnsproject/internal/services/dict"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/miekg/dns"
	flag "github.com/spf13/pflag"
)

var (
	lo = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	ko = koanf.New(".")

	// Version of the build injected at build time.
	buildString = "unknown"
)

const HELP_TTL = 86400

const SIGUNUSED = syscall.Signal(0x1f)

func initConfig() {
	// Register --help handler.
	f := flag.NewFlagSet("config", flag.ContinueOnError)
	f.Usage = func() {
		fmt.Println(f.FlagUsages())
		os.Exit(0)
	}
	f.StringSlice("config", []string{"config.toml"}, "path to one or more TOML config files to load in order")
	f.Bool("version", false, "show build version")
	f.Parse(os.Args[1:])

	// Display version.
	if ok, _ := f.GetBool("version"); ok {
		fmt.Println(buildString)
		os.Exit(0)
	}

	// Read the config files.
	cFiles, _ := f.GetStringSlice("config")
	for _, f := range cFiles {
		lo.Printf("reading config: %s", f)
		if err := ko.Load(file.Provider(f), toml.Parser()); err != nil {
			lo.Printf("error reading config: %v", err)
		}
	}

	ko.Load(posflag.Provider(f, ".", ko), nil)
}

func saveSnapshot(h *handlers) {
	interruptSignal := make(chan os.Signal)
	signal.Notify(interruptSignal,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGINT,
		SIGUNUSED, // SIGUNUSED, can be used to avoid shutting down the app.
	)

	// On receiving an OS signal, iterate through services and
	// dump their snapshots to the disk if available.
	for {
		select {
		case i := <-interruptSignal:
			lo.Printf("received SIGNAL: `%s`", i.String())

			for name, s := range h.services {
				if !ko.Bool(name+".enabled") || !ko.Bool(name+".snapshot_enabled") {
					continue
				}

				b, err := s.Dump()
				if err != nil {
					lo.Printf("error generating %s snapshot: %v", name, err)
				}

				if b == nil {
					continue
				}

				filePath := ko.MustString(name + ".snapshot_file")
				lo.Printf("saving %s snapshot to %s", name, filePath)
				if err := os.WriteFile(filePath, b, 0644); err != nil {
					lo.Printf("error writing %s snapshot: %v", name, err)
				}
			}

			if i != SIGUNUSED {
				os.Exit(0)
			}
		}
	}
}

func loadSnapshot(service string) []byte {
	if !ko.Bool(service + ".snapshot_enabled") {
		return nil
	}

	filePath := ko.MustString(service + ".snapshot_file")

	b, err := os.ReadFile(filePath)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return nil
		}
		lo.Printf("error reading snapshot file %s: %v", filePath, err)
		return nil
	}

	return b
}

func main() {
	initConfig()

	var (
		h = &handlers{
			services: make(map[string]Service),
			domain:   ko.MustString("server.domain"),
		}
		//ge  *geo.Geo
		mux = dns.NewServeMux()

		help = [][]string{}
	)

	// Random number generator.
	if ko.Bool("rand.enabled") {
		// seed the RNG:
		rand.Seed(time.Now().Unix())

		n := random.New()
		h.register("rand", n, mux)

		help = append(help, []string{"generate random numbers", "dig @%s -p %s 1-100.rand "})
	}

	// PI.
	if ko.Bool("pi.enabled") {
		mux.HandleFunc("pi.", h.handlePi)

		help = append(help, []string{"return digits of Pi as TXT or A or AAAA record.", "dig @%s -p @%s pi "})
	}

	// Rolling dice
	if ko.Bool("dice.enabled") {
		n := dice.New()
		h.register("dice", n, mux)

		help = append(help, []string{"roll dice", "dig @%s -p %s 1d6.dice "})
	}

	// FX currency conversion.
	if ko.Bool("fx.enabled") {
		f := fx.New(fx.Opt{
			RefreshInterval: ko.MustDuration("fx.refresh_interval"),
		})

		// Load snapshot?
		if b := loadSnapshot("fx"); b != nil {
			if err := f.Load(b); err != nil {
				lo.Printf("error reading fx snapshot: %v", err)
			}
		}

		h.register("fx", f, mux)

		help = append(help, []string{"convert currency rates", "dig %s -p %s 99USD-INR.fx "})
	}

	// Dictionary.
	if ko.Bool("dict.enabled") {
		d := dict.New(dict.Opt{
			WordNetPath: ko.MustString("dict.wordnet_path"),
			MaxResults:  ko.MustInt("dict.max_results"),
		})
		h.register("dict", d, mux)

		help = append(help, []string{"get the definition of an English word, powered by WordNet(R).", "dig @%s -p %s fun.dict"})
	}

	// IP echo.
	if ko.Bool("ip.enabled") {
		mux.HandleFunc("ip.", h.handleEchoIP)

		help = append(help, []string{"get your host's requesting IP.", "dig @%s -p %s ip"})
	}

	// Get the Port
	_, port, err := net.SplitHostPort(ko.MustString("server.address"))
	if err != nil {
		port = strings.TrimPrefix(ko.MustString("server.address"), ":")
	}

	

	// Prepare the static help response for the `help` query.
	for _, l := range help {
		r, err := dns.NewRR(fmt.Sprintf("help. %d TXT \"%s\" \"%s\"", HELP_TTL, l[0], fmt.Sprintf(l[1], h.domain, port)))
		if err != nil {
			lo.Fatalf("error preparing: %v", err)
		}

		h.help = append(h.help, r)
	}

	mux.HandleFunc("help.", h.handleHelp)
	mux.HandleFunc(".", (h.handleDefault))

	// Start the snapshot listener.
	go saveSnapshot(h)

	// Start the server.
	server := &dns.Server{
		Addr:    ko.MustString("server.address"),
		Net:     "udp",
		Handler: mux,
	}

	lo.Println("listening on ", ko.String("server.address"))
	if err := server.ListenAndServe(); err != nil {
		lo.Fatalf("error starting server: %v", err)
	}
	defer server.Shutdown()
}
