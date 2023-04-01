package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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
		fmt.Println(getProgramVersion())
	}

	mqttURLVar, present := os.LookupEnv("MQTT_URL")

	if !present {
		fmt.Fprintf(os.Stderr, "Error: Required MQTT_URL not present\n")
		os.Exit(1)
	}

	mqttURL, err := url.Parse(mqttURLVar)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	mqttOptions := mqtt.NewClientOptions().
		AddBroker(mqttURL.String()).
		SetClientID(getProgramName()).
		SetUsername(mqttURL.User.Username())

	password, isSet := mqttURL.User.Password()

	if isSet {
		mqttOptions.SetPassword(password)
	}

	mqtt.ERROR = log.New(os.Stderr, "", 0)
	mqttClient := mqtt.NewClient(mqttOptions)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		fmt.Fprintf(os.Stderr, "Error: Could not connect to MQTT: %s\n", token.Error())
		os.Exit(1)
	}

	if opts.Verbose {
		fmt.Printf("Connected to MQTT at %s\n", mqttURL.String())
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		alert := webhook.Message{}
		err := json.NewDecoder(r.Body).Decode(&alert)

		if err != nil {
			log.Printf("Could not decode alert: %v", err)
			http.Error(w, "Could not decode alert", http.StatusInternalServerError)
			return
		}

		for _, a := range alert.Alerts {
			message := struct {
				Name     string    `json:"name"`
				URL      string    `json:"url"`
				Status   string    `json:"status"`
				StartsAt time.Time `json:"startsAt"`
				EndsAt   time.Time `json:"endsAt"`
			}{
				Name:     a.Labels["alertname"],
				URL:      a.GeneratorURL,
				Status:   a.Status,
				StartsAt: a.StartsAt,
				EndsAt:   a.EndsAt,
			}

			messageJSON, err := json.Marshal(message)

			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: marshalling the MQTT message failed: %s\n", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			topicPrefix := strings.Replace(mqttURL.Path, "/", "", 1)
			topic := fmt.Sprintf("%s/%s", topicPrefix, a.Labels["alertname"])

			if token := mqttClient.Publish(topic, 0, false, messageJSON); token.Wait() && token.Error() != nil {
				fmt.Fprintf(os.Stderr, "Error: publishing the MQTT message failed: %s\n", token.Error())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			if opts.Verbose {
				fmt.Printf("Sent to %s: %s\n", topic, messageJSON)
			}
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, http.StatusText(http.StatusCreated))
	})

	if opts.Verbose {
		log.Printf("Starting to listen at http://%v\n", opts.HttpBindAddress)
	}

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
	return fmt.Sprintf("%s %s (%s), built on %s", getProgramName(), version, commit, date)
}
