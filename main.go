
package main

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.Info("creating kafka producer")

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

		if kafkaSslCACertFile != "" {
		    kafkaConfig["ssl.ca.location"] = kafkaSslCACertFile
		}
	}

	//创建生产者
	producer, err := kafka.NewProducer(&kafkaConfig)

	if err != nil {
		logrus.WithError(err).Fatal("couldn't create kafka producer")
	}


	//创建topic
	if kafkaTopic!=""{
		createTopic(producer,[]string{kafkaTopic})
	}

	r := gin.New()

	r.Use(ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true), gin.Recovery())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "OK"}) })
	if basicauth {
		authorized := r.Group("/", gin.BasicAuth(gin.Accounts{
			basicauthUsername: basicauthPassword,
		}))
		authorized.POST("/receive", receiveHandler(producer))
	} else {
		r.POST("/receive", receiveHandler(producer))
	}

	logrus.Fatal(r.Run())
}
