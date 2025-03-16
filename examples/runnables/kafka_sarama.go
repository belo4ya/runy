package runnables

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type KafkaSaramaConsumer struct {
	client  sarama.ConsumerGroup
	conf    KafkaSaramaConsumerConfig
	handler sarama.ConsumerGroupHandler
	wg      sync.WaitGroup
}

type KafkaSaramaConsumerConfig struct {
	Brokers      []string
	GroupID      string
	Topics       []string
	SaramaConfig *sarama.Config
}

func NewKafkaSaramaConsumer(conf KafkaSaramaConsumerConfig, handler sarama.ConsumerGroupHandler) (*KafkaSaramaConsumer, error) {
	if conf.SaramaConfig == nil {
		conf.SaramaConfig = sarama.NewConfig()
		conf.SaramaConfig.Consumer.Return.Errors = true
	}

	client, err := sarama.NewConsumerGroup(conf.Brokers, conf.GroupID, conf.SaramaConfig)
	if err != nil {
		return nil, fmt.Errorf("create consumer group: %w", err)
	}

	return &KafkaSaramaConsumer{
		client:  client,
		conf:    conf,
		handler: handler,
	}, nil
}

func (c *KafkaSaramaConsumer) Start(ctx context.Context) error {
	// Use a separate cancellable context for the consumer loop
	consumerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			log.Printf("kafka consumer starting consumption from topics: %v, group: %s", c.conf.Topics, c.conf.GroupID)
			if err := c.client.Consume(consumerCtx, c.conf.Topics, c.handler); err != nil {
				if !errors.Is(err, sarama.ErrClosedConsumerGroup) {
					errCh <- fmt.Errorf("consume: %w", err)
					return
				}
			}

			if consumerCtx.Err() != nil {
				return
			}
		}
	}()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for err := range c.client.Errors() {
			log.Printf("kafka consumer error: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("shutting down kafka consumer")
		cancel()

		// Close the consumer and wait for goroutines to finish
		if err := c.client.Close(); err != nil {
			return fmt.Errorf("failed to close consumer: %w", err)
		}

		c.wg.Wait()
		return nil
	case err := <-errCh:
		return err
	}
}

type KafkaSaramaConsumerHandler struct {
	ready    chan bool
	handler  func(message *sarama.ConsumerMessage) error
	shutdown chan struct{}
}

func NewKafkaSaramaConsumerHandler(handler func(message *sarama.ConsumerMessage) error) *KafkaSaramaConsumerHandler {
	return &KafkaSaramaConsumerHandler{
		ready:    make(chan bool),
		handler:  handler,
		shutdown: make(chan struct{}),
	}
}

func (h *KafkaSaramaConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	close(h.ready)
	return nil
}

func (h *KafkaSaramaConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *KafkaSaramaConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := h.handler(message); err != nil {
				log.Printf("Error processing message: %v", err)
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		case <-h.shutdown:
			return nil
		}
	}
}

func (h *KafkaSaramaConsumerHandler) Stop() {
	close(h.shutdown)
}
