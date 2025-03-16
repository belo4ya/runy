package runnables

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaFranzConsumer struct {
	client  *kgo.Client
	conf    KafkaFranzConsumerConfig
	handler func(*kgo.Record) error
	wg      sync.WaitGroup
}

type KafkaFranzConsumerConfig struct {
	Brokers []string
	GroupID string
	Topics  []string
	Config  []kgo.Opt
}

func NewKafkaFranzConsumer(conf KafkaFranzConsumerConfig, handler func(*kgo.Record) error) (*KafkaFranzConsumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(conf.Brokers...),
		kgo.ConsumerGroup(conf.GroupID),
		kgo.ConsumeTopics(conf.Topics...),
		kgo.DisableAutoCommit(),
	}

	if conf.Config != nil {
		opts = append(opts, conf.Config...)
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &KafkaFranzConsumer{
		client:  client,
		conf:    conf,
		handler: handler,
	}, nil
}

func (c *KafkaFranzConsumer) Start(ctx context.Context) error {
	log.Printf("franz kafka consumer starting consumption from topics: %v, group: %s",
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
				fetches := c.client.PollFetches(ctx)
				if fetches.IsClientClosed() {
					return
				}

				if errs := fetches.Errors(); len(errs) > 0 {
					for _, err := range errs {
						log.Printf("kafka fetch error: topic %s, partition %d: %v",
							err.Topic, err.Partition, err.Err)
					}
					continue
				}

				fetches.EachRecord(func(record *kgo.Record) {
					if err := c.handler(record); err != nil {
						log.Printf("error processing record: %v", err)
					} else {
						c.client.MarkCommitRecords(record)
					}
				})

				if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
					log.Printf("error committing offsets: %v", err)
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down franz kafka consumer")
		c.client.Close()
		c.wg.Wait()
		return nil
	case err := <-errCh:
		return err
	}
}
