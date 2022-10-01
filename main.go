package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/prometheus/alertmanager/notify/webhook"
)

type options struct {
	Version         bool   `short:"V" long:"version" description:"Print version information and exit"`
	Verbose         bool   `short:"v" long:"verbose" description:"Print verbose information"`
	HttpBindAddress string `short:"b" long:"bind" description:"Address to bind the HTTP control server to" default:"localhost:8031"`
}

// ldflags will be set by goreleaser
var version = "vDEV"
var commit = "NONE"
var date = "UNKNOWN"

var opts options

func main() {
	log.SetFlags(0) // no timestamp etc. - we have systemd's timestamps in the log anyway

	_, err := flags.Parse(&opts)

	if err != nil {
		os.Exit(1)
	}

	if opts.Version {
		log.Println(getProgramVersion())
		os.Exit(0)
	}

	if opts.Verbose {
		log.Printf("Starting with options: %v\n", opts)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		message := webhook.Message{}
		err := json.NewDecoder(r.Body).Decode(&message)

		if err != nil {
			log.Printf("Could not switch to tab: %v", err)
			http.Error(w, "Could not switch to tab", http.StatusInternalServerError)
			return
		}

		for _, a := range message.Alerts {
			log.Printf("%v %v", a.Labels["alertname"], a.Status)
		}

		w.WriteHeader(http.StatusCreated)
	})

	log.Printf("%v starting at http://%v\n", getProgramVersion(), opts.HttpBindAddress)
	log.Fatal(http.ListenAndServe(opts.HttpBindAddress, nil))
}

func getProgramName() string {
	path, err := os.Executable()

	if err != nil {
		fmt.Fprintln(os.Stderr, "Warning: Could not determine program name; using 'unknown'.")
		return "unknown"
	}

	return filepath.Base(path)
}

func getProgramVersion() string {
	return fmt.Sprintf("%s %s (%s), built on %s\n", getProgramName(), version, commit, date)
}
