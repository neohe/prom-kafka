package main

import (
    "io/ioutil"
    "net/http"

    "github.com/Shopify/sarama"
    "github.com/gin-gonic/gin"
    "github.com/gogo/protobuf/proto"
    "github.com/golang/snappy"
    "github.com/prometheus/prometheus/prompb"
    "github.com/sirupsen/logrus"
)

func receiveHandler(producer sarama.AsyncProducer, serializer Serializer) func(c *gin.Context) {
    return func(c *gin.Context) {

        httpRequestsTotal.Add(float64(1))

        compressed, err := ioutil.ReadAll(c.Request.Body)
        if err != nil {
            c.AbortWithStatus(http.StatusInternalServerError)
            logrus.WithError(err).Error("couldn't read body")
            return
        }

        reqBuf, err := snappy.Decode(nil, compressed)
        if err != nil {
            c.AbortWithStatus(http.StatusBadRequest)
            logrus.WithError(err).Error("couldn't decompress body")
            return
        }

        var req prompb.WriteRequest
        if err := proto.Unmarshal(reqBuf, &req); err != nil {
            c.AbortWithStatus(http.StatusBadRequest)
            logrus.WithError(err).Error("couldn't unmarshal body")
            return
        }

        metricsPerTopic, err := processWriteRequest(&req)
        if err != nil {
            c.AbortWithStatus(http.StatusInternalServerError)
            logrus.WithError(err).Error("couldn't process write request")
            return
        }

        go func() {
            for err := range producer.Errors() {
                logrus.WithError(err).Error("failed to produce")
            }
        }()

        for topic, metrics := range metricsPerTopic {
            for _, metric := range metrics {
                producer.Input() <- &sarama.ProducerMessage{
                    Topic: topic,
                    Value: sarama.ByteEncoder(metric),
                }
            }
        }
    }
}
