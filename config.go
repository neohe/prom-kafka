package main

import (
    "os"
    "strings"
    "text/template"

    "github.com/sirupsen/logrus"
)

var (
    kafkaBrokerList   = []string{"kafka1:9092", "kafka2:9092", "kafka3:9092"}
    kafkaTopic        = "metrics"
    topicTemplate     *template.Template
    basicauth         = false
    basicauthUsername = ""
    basicauthPassword = ""
    serializer        Serializer
)

func init() {
    logrus.SetFormatter(&logrus.JSONFormatter{})
    logrus.SetOutput(os.Stdout)

    if value := os.Getenv("LOG_LEVEL"); value != "" {
        logrus.SetLevel(parseLogLevel(value))
    }

    if value := os.Getenv("KAFKA_BROKER_LIST"); value != "" {
        kafkaBrokerList = strings.Split(value, ",")
    }

    if value := os.Getenv("KAFKA_TOPIC"); value != "" {
        kafkaTopic = value
    }

    if value := os.Getenv("BASIC_AUTH_USERNAME"); value != "" {
        basicauth = true
        basicauthUsername = value
    }

    if value := os.Getenv("BASIC_AUTH_PASSWORD"); value != "" {
        basicauthPassword = value
    }

    var err error
    serializer, err = parseSerializationFormat(os.Getenv("SERIALIZATION_FORMAT"))
    if err != nil {
        logrus.WithError(err).Fatalln("couldn't create a metrics serializer")
    }

    topicTemplate, err = parseTopicTemplate(kafkaTopic)
    if err != nil {
        logrus.WithError(err).Fatalln("couldn't parse the topic template")
    }
}

func parseLogLevel(value string) logrus.Level {
    level, err := logrus.ParseLevel(value)

    if err != nil {
        logrus.WithField("log-level-value", value).Warningln("invalid log level from env var, using info")
        return logrus.InfoLevel
    }

    return level
}

func parseSerializationFormat(value string) (Serializer, error) {
    switch value {
    case "json":
        return NewJSONSerializer()
    default:
        logrus.WithField("serialization-format-value", value).Warningln("invalid serialization format, using json")
        return NewJSONSerializer()
    }
}

func parseTopicTemplate(tpl string) (*template.Template, error) {
    funcMap := template.FuncMap{
        "replace": func(old, new, src string) string {
            return strings.Replace(src, old, new, -1)
        },
        "substring": func(start, end int, s string) string {
            if start < 0 {
                start = 0
            }
            if end < 0 || end > len(s) {
                end = len(s)
            }
            if start >= end {
                panic("template function - substring: start is bigger (or equal) than end. That will produce an empty string.")
            }
            return s[start:end]
        },
    }
    return template.New("topic").Funcs(funcMap).Parse(tpl)
}
