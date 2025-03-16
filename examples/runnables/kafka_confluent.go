package runnables

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type KafkaConfluentConsumer struct {
	consumer *kafka.Consumer
	conf     KafkaConfluentConsumerConfig
	handler  func(*kafka.Message) error
	wg       sync.WaitGroup
}

type KafkaConfluentConsumerConfig struct {
	BootstrapServers string
	GroupID          string
	Topics           []string
	AutoOffsetReset  string
	PollTimeout      time.Duration
	Config           *kafka.ConfigMap
}

func NewKafkaConfluentConsumer(conf KafkaConfluentConsumerConfig, handler func(*kafka.Message) error) (*KafkaConfluentConsumer, error) {
	if conf.PollTimeout == 0 {
		conf.PollTimeout = 100 * time.Millisecond
	}

	if conf.AutoOffsetReset == "" {
		conf.AutoOffsetReset = "earliest"
	}

	configMap := &kafka.ConfigMap{
		"bootstrap.servers":  conf.BootstrapServers,
		"group.id":           conf.GroupID,
		"auto.offset.reset":  conf.AutoOffsetReset,
		"enable.auto.commit": true,
	}

	if conf.Config != nil {
		for k, v := range *conf.Config {
			if err := configMap.SetKey(k, v); err != nil {
				return nil, fmt.Errorf("set config: %w", err)
			}
		}
	}

	consumer, err := kafka.NewConsumer(configMap)
	if err != nil {
		return nil, fmt.Errorf("create consumer: %w", err)
	}

	return &KafkaConfluentConsumer{
		consumer: consumer,
		conf:     conf,
		handler:  handler,
	}, nil
}

func (c *KafkaConfluentConsumer) Start(ctx context.Context) error {
	if err := c.consumer.SubscribeTopics(c.conf.Topics, nil); err != nil {
		return fmt.Errorf("subscribe topics: %w", err)
	}

	log.Printf("confluent kafka consumer starting consumption from topics: %v, group: %s",
		c.conf.Topics, c.conf.GroupID)

	errCh := make(chan error, 1)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				ev := c.consumer.Poll(int(c.conf.PollTimeout.Milliseconds()))
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					if err := c.handler(e); err != nil {
						log.Printf("error processing message: %v", err)
					}
				case kafka.Error:
					if e.IsFatal() {
						errCh <- fmt.Errorf("consumer error: %w", e)
						return
					}
					log.Printf("non-fatal consumer error: %v", e)
				default:
					// Ignore other event types
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down confluent kafka consumer")

		err := c.consumer.Close()
		c.wg.Wait()

		if err != nil {
			return fmt.Errorf("close consumer: %w", err)
		}
		return nil
	case err := <-errCh:
		return err
	}
}
