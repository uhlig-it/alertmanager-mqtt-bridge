# Alertmanager MQTT Bridge

This is a Webhook server converting [Prometheus Alertmanager webhook messages](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config) into MQTT messages.

The MQTT topic is taken from the `MQTT_URL`'s path, and the alert name is appended.

Example: With `MQTT_URL=mqtts://user:password@mqtt.example.com/alerts/grafana`, an alert named "My Alert" is published to `/alerts/grafana/My Alert`.

# Deployment

```command
$ (cd deployment && ansible-playbook playbook.yml)
```

# Development

* Iterate with [`entr`](https://eradman.com/entrproject/):

  ```command
  $ find . -name '*.go' -or -name '*.tmpl' -type f | entr -r go run . --verbose
  ```

* Listen to MQTT messages:

  ```command
  $ mosquitto_sub --url 'mqtts://user:password@mqtt.example.com/alerts/grafana/#' -F %J | jq
  ```

* Post a fixture to the running server:

  ```command
  $ curl -v localhost:8031 -d @fixtures/alert.json
  ```

Check `.tmuxinator.yml` for an ready-to-launch configuration of the above.
