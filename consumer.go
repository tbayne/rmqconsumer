package rmqconsumer

import (
	"crypto/tls"
	"fmt"

	log "github.com/cihub/seelog"
	"github.com/streadway/amqp"
)

// The Consumer structure holds our connection to Rabbit
type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
}

// ConsumerHandler defines our consumer handler
type ConsumerHandler func(<-chan amqp.Delivery, chan error)

// Connects to RabbitMQ, starts the provided handler as a Go routine, and
// returns an initialized Consumer structure.

// NewConsumer returns a Consumer structure and starts the ConsumerHandler
func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, ch ConsumerHandler, sslcfg *tls.Config) (*Consumer, error) {
	log.Trace("NewConsumer() start")
	c := &Consumer{
		conn:    nil,
		channel: nil,
		tag:     ctag,
		done:    make(chan error),
	}

	var err error

	if sslcfg != nil {
		fmt.Println("sslcfg != nil, so attempting SSL connection")
		c.conn, err = amqp.DialTLS(amqpURI, sslcfg)

	} else {
		//Trace.Printf("dialing %q", amqpURI)
		c.conn, err = amqp.Dial(amqpURI)

	}

	if err != nil {
		log.Error("Error connecting to server (", amqpURI, ") %s", err)
		return nil, fmt.Errorf("Dial: %s", err)
	}
	//c.publishers = publishers
	log.Trace("Connected")

	go func() {
		log.Trace("Go func(): waiting for NotifyClose")
		msg := fmt.Sprintf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
		log.Trace(msg)
		log.Trace("Go Function - last line")
	}()

	//Trace.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		log.Error("Error creating channel ", err)
		return nil, fmt.Errorf("Channel: %s", err)
	}
	log.Trace("Channel created")
	//Trace.Printf("got Channel, declaring Exchange (%q)", exchange)
	if err = c.channel.ExchangeDeclare(
		exchange,     // name of the exchange
		exchangeType, // type
		true,         // durable
		false,        // delete when complete
		false,        // internal
		false,        // noWait
		nil,          // arguments
	); err != nil {
		log.Error("Error declaring exchange: ", err)
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}
	log.Trace("Exchange Declared")

	//Trace.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		queueName, // name of the queue
		true,      // durable
		false,     // delete when usused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		log.Error("Error declaring queue: ", err)
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}
	log.Trace("Queue declared")
	//Trace.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
	//	queue.Name, queue.Messages, queue.Consumers, key)

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		key,        // bindingKey
		exchange,   // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		log.Error("Error binding queue to exchange:", err)
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	log.Trace("Queue bound to exchange")
	//Trace.Printf("Queue bound to Exchange, starting Consume (consumer tag %q)", c.tag)
	deliveries, err := c.channel.Consume(
		queue.Name, // name
		c.tag,      // consumerTag,
		false,      // noAck
		false,      // exclusive
		false,      // noLocal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		log.Error("Error consuming messages: ", err)
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}
	log.Trace("Consume started, firing  handle() as a Go Routine")
	// This fires off our handler as a background task to handle incoming messages.
	go ch(deliveries, c.done)

	return c, nil
}

// Shutdown closes down the consumer and waits for the handler to exit
func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	log.Trace("Start")

	if err := c.channel.Cancel(c.tag, true); err != nil {
		log.Error("Consumer cancel failed: ", err)
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		log.Error("AMQP connection close error: ", err)
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	// wait for handle() to exit
	return <-c.done
}
