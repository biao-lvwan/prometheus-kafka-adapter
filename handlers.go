package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/snappy"

	"github.com/gogo/protobuf/proto"
	"github.com/prometheus/prometheus/prompb"
)

func createTopic(producer *kafka.Producer,topics []string){
	brokers := strings.Split(kafkaBrokerList, ",")
	numBrokers := len(brokers) // 获取 broker 数量
	numPartition := 30
	replicationFactor:= 3
	if numPartitionValue !=""{
		numPartition, _ = strconv.Atoi(numPartitionValue)
	}else{
		if numBrokers == 1 {
			numPartition = 8
		}
	}
	if replicationFactorValue !=""{
		replicationFactor, _ = strconv.Atoi(replicationFactorValue)
	}else{
		if numBrokers == 1 {
			replicationFactor =1
		}
	}

	// 创建一个新的AdminClient。
	a, err := kafka.NewAdminClientFromProducer(producer)
	if err != nil {
		fmt.Printf("Failed to create Admin client: %s\n", err)
		os.Exit(1)
	}

	// Contexts 用于中止或限制时间量
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//在集群上创建主题。
	//设置管理员选项以等待操作完成（或最多60秒）
	maxDur, err := time.ParseDuration("60s")
	if err != nil {
		logrus.WithError(err).Fatal("ParseDuration(60s)")
	}
	//创建主题因素
	topicSpecifications := []kafka.TopicSpecification{}
	metadata, _ := a.GetMetadata(nil, true, 5000)
	if err != nil {
		logrus.WithError(err).Fatal("query topic error:")
	}
	createTopics :=[]string{}
	if metadata==nil{
		createTopics = append(topics)
	}else{
		for _,  value:= range topics  {
			flag :=true
			for _,  t:= range metadata.Topics {
				if t.Topic == value {
					flag =false
					break
				}
			}
			if flag{
				createTopics= append(createTopics, value)
			}
		}
	}
	logrus.Infof("获取节点信息，连接数:"+strconv.Itoa(numBrokers)+"分区数:"+strconv.Itoa(numPartition)+"副本数:"+strconv.Itoa(replicationFactor)+"自定义分区数:"+numPartitionValue+"自定义副本数:"+replicationFactorValue)
	for _, value := range createTopics {
		specification := kafka.TopicSpecification{
			Topic:             value,
			NumPartitions:     numPartition,
			ReplicationFactor: replicationFactor}
		topicSpecifications= append(topicSpecifications,specification )
	}
	results, err := a.CreateTopics(
		ctx,
		//通过提供的TopicSpecification结构，可以同时创建多个主题
		topicSpecifications,
		// Admin options
		kafka.SetAdminOperationTimeout(maxDur))
	if err != nil {
		fmt.Printf("Failed to create topic: %v\n", err)
		os.Exit(1)
	}
	// Print results
	for _, result := range results {
		fmt.Printf("%s\n", result)
	}
	a.Close()
}

func receiveHandler(producer *kafka.Producer, serializer Serializer) func(c *gin.Context) {
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
        //获取所有的topic手工创建
		/*topics :=[]string{}
		for topic, _ := range metricsPerTopic {
			topics= append(topics, topic)
		}
		//检查是否创建topic,可用动态创建topic，但是会损耗效率，关闭此功能
		createTopic(kafkaConfig,topics)*/
		for topic, metrics := range metricsPerTopic {

			t := topic
			part := kafka.TopicPartition{
				Partition: kafka.PartitionAny,
				Topic:     &t,
			}
			for _, metric := range metrics {
				objectsWritten.Add(float64(1))
				err := producer.Produce(&kafka.Message{
					TopicPartition: part,
					Value:          metric,
				}, nil)

				if err != nil {
					objectsFailed.Add(float64(1))
					c.AbortWithStatus(http.StatusInternalServerError)
					logrus.WithError(err).Debug(fmt.Sprintf("Failing metric %v", metric))
					logrus.WithError(err).Error(fmt.Sprintf("couldn't produce message in kafka topic %v", topic))
					return
				}
			}
		}

	}
}
