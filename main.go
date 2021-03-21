package main

import (
    "github.com/confluentinc/confluent-kafka-go/kafka"
    "github.com/gin-contrib/logger"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/common/log"
    "github.com/sirupsen/logrus"
)

func main() {
    log.Info("creating kafka producer")

    kafkaConfig := kafka.ConfigMap{
        "bootstrap.servers":   kafkaBrokerList,
        "compression.codec":   kafkaCompression,
        "batch.num.messages":  kafkaBatchNumMessages,
        "go.batch.producer":   true,  // Enable batch producer (for increased performance).
        "go.delivery.reports": false, // per-message delivery reports to the Events() channel
    }

    if kafkaSslClientCertFile != "" && kafkaSslClientKeyFile != "" && kafkaSslCACertFile != "" {
        if kafkaSecurityProtocol == "" {
            kafkaSecurityProtocol = "ssl"
        }

        if kafkaSecurityProtocol != "ssl" && kafkaSecurityProtocol != "sasl_ssl" {
            logrus.Fatal("invalid config: kafka security protocol is not ssl based but ssl config is provided")
        }

        kafkaConfig["security.protocol"] = kafkaSecurityProtocol
        kafkaConfig["ssl.ca.location"] = kafkaSslCACertFile              // CA certificate file for verifying the broker's certificate.
        kafkaConfig["ssl.certificate.location"] = kafkaSslClientCertFile // Client's certificate
        kafkaConfig["ssl.key.location"] = kafkaSslClientKeyFile          // Client's key
        kafkaConfig["ssl.key.password"] = kafkaSslClientKeyPass          // Key password, if any.
    }

    if kafkaSaslMechanism != "" && kafkaSaslUsername != "" && kafkaSaslPassword != "" {
        if kafkaSecurityProtocol != "sasl_ssl" && kafkaSecurityProtocol != "sasl_plaintext" {
            logrus.Fatal("invalid config: kafka security protocol is not sasl based but sasl config is provided")
        }

        kafkaConfig["security.protocol"] = kafkaSecurityProtocol
        kafkaConfig["sasl.mechanism"] = kafkaSaslMechanism
        kafkaConfig["sasl.username"] = kafkaSaslUsername
        kafkaConfig["sasl.password"] = kafkaSaslPassword
    }

    producer, err := kafka.NewProducer(&kafkaConfig)
    if err != nil {
        logrus.WithError(err).Fatal("couldn't create kafka producer")
    }

    r := gin.New()

    r.Use(logger.SetLogger(), gin.Recovery())

    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    if basicauth {
        authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
            basicauthUsername: basicauthPassword,
        }))
        authorized.POST("/write", receiveHandler(producer, serializer))
    } else {
        r.POST("/write", receiveHandler(producer, serializer))
    }

    logrus.Fatal(r.Run())
}
