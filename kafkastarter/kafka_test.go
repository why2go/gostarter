package kafkastarter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/stretchr/testify/assert"
)

func TestKafka(t *testing.T) {
	const TOPIC_NAME = "my_topic_2"

	brokers := []string{
		"localhost:63900",
		"localhost:63901",
		"localhost:63902",
	}

	t.Run("Admin", func(t *testing.T) {
		var err error
		config := sarama.NewConfig()
		admin, err := sarama.NewClusterAdmin(brokers, config)
		assert.Empty(t, err)
		defer admin.Close()

		topics, err := admin.ListTopics()
		assert.Empty(t, err)

		for tn, td := range topics {
			fmt.Printf("topic name: %v\n", tn)
			b, err := json.Marshal(td)
			assert.Empty(t, err)
			fmt.Printf("topic details: %v\n", string(b))
			fmt.Println("==============")
		}

	})

	// 获取当前偏移量
	t.Run("CurOffset", func(t *testing.T) {
	})

	// 向topic中推送1000条消息
	t.Run("BulkPush", func(t *testing.T) {
		config := sarama.NewConfig()
		producer, err := sarama.NewAsyncProducer(brokers, config)
		assert.Empty(t, err)
		defer producer.Close()

		wg := sync.WaitGroup{}

		// 监听错误
		go func() {
			for err := range producer.Errors() {
				fmt.Printf("err: %v\n", err)
			}
		}()

		// 发送消息
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 1000; i++ {
				producer.Input() <- &sarama.ProducerMessage{
					Topic: TOPIC_NAME,
					Value: sarama.StringEncoder(strconv.Itoa(i)),
				}
			}
			time.Sleep(time.Second)
		}()

		wg.Wait()
	})

	// 从offset第200条开始消费
	t.Run("Offset", func(t *testing.T) {
		config := sarama.NewConfig()
		config.Consumer.Offsets.Initial = sarama.OffsetNewest

		c01, err := sarama.NewConsumer(brokers, config)
		assert.Empty(t, err)
		defer c01.Close()

		partitions, err := c01.Partitions(TOPIC_NAME)
		assert.Empty(t, err)
		fmt.Printf("partitions: %v\n", partitions)

		pc, err := c01.ConsumePartition(TOPIC_NAME, 0, 7)
		assert.Empty(t, err)

		for i := 0; i < 10; i++ {
			msg := <-pc.Messages()

			fmt.Printf("topic: %v, partition: %v, offset: %v, value: %v, timestamp: %v\n",
				msg.Topic, msg.Partition, msg.Offset, string(msg.Value), msg.Timestamp)
			time.Sleep(time.Second)
		}
	})

	// 从offset第500条开始消费
	t.Run("ConsumerGroup", func(t *testing.T) {
		var err error
		config := sarama.NewConfig()
		cg01, err := sarama.NewConsumerGroup(brokers, "csg-02", config)
		assert.Empty(t, err)

		go func() {
			for err := range cg01.Errors() {
				fmt.Printf("cg01 err: %v\n", err)
			}
		}()

		defer func() {
			err := cg01.Close()
			assert.Empty(t, err)
		}()

		h := MyHandler{Topic: TOPIC_NAME, Offset: 7}

		err = cg01.Consume(context.Background(), []string{TOPIC_NAME}, h)
		assert.Empty(t, err)
	})
}

type MyHandler struct {
	Topic  string
	Offset int64
}

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (h MyHandler) Setup(cgs sarama.ConsumerGroupSession) error {
	// cgs.ResetOffset(h.Topic, 0, h.Offset, "")
	// cgs.Commit()
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
// but before the offsets are committed for the very last time.
func (h MyHandler) Cleanup(cgs sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
// Once the Messages() channel is closed, the Handler must finish its processing
// loop and exit.
func (h MyHandler) ConsumeClaim(cgs sarama.ConsumerGroupSession, cgc sarama.ConsumerGroupClaim) error {
	fmt.Printf("cgc.InitialOffset(): %v\n", cgc.InitialOffset())
	count := 1
	for msg := range cgc.Messages() {
		fmt.Printf("topic: %v, partition: %v, offset: %v, value: %v, timestamp: %v\n",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Value), msg.Timestamp)
		cgs.MarkMessage(msg, "")
		time.Sleep(time.Second)
		count++
		if count > 10 {
			break
		}
	}
	return nil
}
