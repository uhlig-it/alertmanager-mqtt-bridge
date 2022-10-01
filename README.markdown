# Alertmanager MQTT Bridge

This is a Webhook server converting [Prometheus Alertmanager webhook messages]() into MQTT messages.

# Deployment

```command
$ cd deployment
$ ansible-playbook playbook.yml
```

# Manual Test

* `go build`
* `scp` the binary to the server
* Set the environment variables:
  - `MQTT_URL` (required) must point to the MQTT server where the alerts will be published to, incl. the topic as URL path (leading slash will be removed)
* Start the server
* Configure AlertManager as external alert manager in Grafana (e.g. 127.0.0.1:9093)
* Check that the notification arrives:

  ```command
  $ mosquitto_sub \
    --url 'mqtt://user:pass@example.com/alerts/+' \
    -F '\e[92m %I %t: \e[96m%p\e[0m'
  ```

# Development

* Use `fresh` for fast iteration (`go get github.com/pilu/fresh`)
* Post a fixture to the running server:

  ```command
  $ curl -v localhost:8031 -d @fixtures/alert.json
  ```
