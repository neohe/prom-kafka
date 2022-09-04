# prom-kafka

Prom-kafka is a service which receives [Prometheus](https://github.com/prometheus) metrics through [`remote_write`](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#remote_write), marshal into JSON and sends them into [Kafka](https://github.com/apache/kafka).

## output

It is able to write JSON messages in a kafka topic.

### JSON

```json
{
  "timestamp": "1970-01-01T00:00:00Z",
  "value": "9876543210",
  "name": "up",

  "labels": {
    "__name__": "up",
    "label1": "value1",
    "label2": "value2"
  }
}
```

`timestamp` and `value` are reserved values, and can't be used as label names. `__name__` is a special label that defines the name of the metric and is copied as `name` to the top level for convenience.

## configuration

### prom-kafka

Prom-kafka listens for metrics coming from Prometheus and sends them to Kafka. This behaviour can be configured with the following environment variables:

- `KAFKA_BROKER_LIST`: defines kafka endpoint and port, defaults to `kafka:9092`.
- `KAFKA_TOPIC`: defines kafka topic to be used, defaults to `metrics`. Could use go template, labels are passed (as a map) to the template: e.g: `metrics.{{ index . "__name__" }}` to use per-metric topic. Two template functions are available: replace (`{{ index . "__name__" | replace "message" "msg" }}`) and substring (`{{ index . "__name__" | substring 0 5 }}`)
- `BASIC_AUTH_USERNAME`: basic auth username to be used for receive endpoint, defaults is no basic auth.
- `BASIC_AUTH_PASSWORD`: basic auth password to be used for receive endpoint, defaults is no basic auth.
- `LOG_LEVEL`: defines log level for [`logrus`](https://github.com/sirupsen/logrus), can be `debug`, `info`, `warn`, `error`, `fatal` or `panic`, defaults to `info`.

### prometheus

Prometheus needs to have a `remote_write` url configured, pointing to the '/write' endpoint of the host and port where the prometheus-kafka-adapter service is running. For example:

```yaml
remote_write:
  - url: "http://prom-kafka:8080/write"
```

## license

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.