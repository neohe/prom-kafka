package main

import (
    "github.com/Shopify/sarama"
    "github.com/gin-contrib/logger"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/prometheus/common/log"
    "github.com/sirupsen/logrus"
)

func main() {
    log.Info("creating kafka producer")

    config := sarama.NewConfig()
    config.Producer.Return.Errors = true
    config.Producer.Partitioner = sarama.NewRandomPartitioner

    client, err := sarama.NewClient(kafkaBrokerList, config)
    if err != nil {
        logrus.WithError(err).Fatal("couldn't not create kafka client")
    }
    defer client.Close()
    producer, err := sarama.NewAsyncProducerFromClient(client)
    if err != nil {
        logrus.WithError(err).Fatal("couldn't create kafka producer")
    }
    defer producer.AsyncClose()

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
