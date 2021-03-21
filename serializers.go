package main

import (
    "bytes"
    "encoding/json"
    "strconv"
    "time"

    "github.com/prometheus/common/model"
    "github.com/prometheus/prometheus/prompb"
    "github.com/sirupsen/logrus"
)

// Serializer represents an abstract metrics serializer
type Serializer interface {
    Marshal(metric map[string]interface{}) ([]byte, error)
}

// Serialize generates the JSON representation for a given Prometheus metric.
func Serialize(s Serializer, req *prompb.WriteRequest) (map[string][][]byte, error) {
    result := make(map[string][][]byte)

    for _, ts := range req.Timeseries {
        labels := make(map[string]string, len(ts.Labels))

        for _, l := range ts.Labels {
            labels[string(model.LabelName(l.Name))] = string(model.LabelValue(l.Value))
        }

        t := topic(labels)

        for _, sample := range ts.Samples {
            epoch := time.Unix(sample.Timestamp/1000, 0).UTC()

            m := map[string]interface{}{
                "timestamp": epoch.Format(time.RFC3339),
                "value":     strconv.FormatFloat(sample.Value, 'f', -1, 64),
                "name":      string(labels["__name__"]),
                "labels":    labels,
            }

            data, err := s.Marshal(m)
            if err != nil {
                logrus.WithError(err).Errorln("couldn't marshal timeseries")
            }

            result[t] = append(result[t], data)
        }
    }

    return result, nil
}

// JSONSerializer represents a metrics serializer that writes JSON
type JSONSerializer struct {
}

func (s *JSONSerializer) Marshal(metric map[string]interface{}) ([]byte, error) {
    return json.Marshal(metric)
}

func NewJSONSerializer() (*JSONSerializer, error) {
    return &JSONSerializer{}, nil
}

func topic(labels map[string]string) string {
    var buf bytes.Buffer
    if err := topicTemplate.Execute(&buf, labels); err != nil {
        return ""
    }
    return buf.String()
}
