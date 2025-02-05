package grmq_test

import (
	"context"
	"testing"
	"time"

	"github.com/integration-system/grmq"
	"github.com/integration-system/grmq/publisher"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/require"
)

func TestPublisher_Publish(t *testing.T) {
	require := require.New(t)

	url := amqpUrl(t)
	ch := amqpChannel(t, url)

	counter := NewObserverCounter()

	pub := publisher.New("", "test")
	unit := grmq.NewPublisher(pub, ch, counter)
	err := unit.Run()
	require.NoError(err)

	err = pub.Publish(context.Background(), &amqp091.Publishing{})
	require.NoError(err)

	declareQueue(t, ch, "test")
	err = pub.Publish(context.Background(), &amqp091.Publishing{})
	require.NoError(err)

	require.EqualValues(1, queueSize(t, url, "test"))
	require.EqualValues(0, counter.publisherError.Load())
	require.EqualValues(0, counter.publisherFlow.Load())

	err = unit.Close()
	require.NoError(err)
}

func TestPublisher_PublishTo(t *testing.T) {
	require := require.New(t)

	url := amqpUrl(t)
	ch := amqpChannel(t, url)

	pub := publisher.New("", "test")
	unit := grmq.NewPublisher(pub, ch, grmq.NoopObserver{})
	err := unit.Run()
	require.NoError(err)

	declareQueue(t, ch, "test2")
	err = pub.PublishTo(context.Background(), "", "test2", &amqp091.Publishing{})
	require.NoError(err)

	require.EqualValues(1, queueSize(t, url, "test2"))

	err = unit.Close()
	require.NoError(err)
}

func TestPublisher_Close(t *testing.T) {
	require := require.New(t)

	url := amqpUrl(t)
	ch := amqpChannel(t, url)

	declareQueue(t, ch, "test")

	pub := publisher.New("", "test")
	unit := grmq.NewPublisher(pub, ch, grmq.NoopObserver{})
	err := unit.Run()
	require.NoError(err)

	err = unit.Close()
	require.NoError(err)

	err = pub.Publish(context.Background(), &amqp091.Publishing{})
	require.Error(err)

	require.EqualValues(0, queueSize(t, url, "test"))
}

func TestPublisherError(t *testing.T) {
	require := require.New(t)

	url := amqpUrl(t)
	ch := amqpChannel(t, url)

	counter := NewObserverCounter()
	pub := publisher.New("undeclared_exchange", "test")
	unit := grmq.NewPublisher(pub, ch, counter)
	err := unit.Run()
	require.NoError(err)

	err = pub.Publish(context.Background(), &amqp091.Publishing{})
	require.NoError(err)

	time.Sleep(500 * time.Millisecond)
	require.EqualValues(1, counter.publisherError.Load())
}
