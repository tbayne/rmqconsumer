# rmqconsumer
A Go based library for consuming messages from a specified RabbitMQ queue

## Build Requriements
 * A Go Development Environment (see: https://golang.org/doc/install)

## Prerequisites:

  * amqp - A go library for accessing amqp servers
   * http://godoc.org/github.com/streadway/amqp
 
 * seelog - a Go logging framework
  * https://github.com/cihub/seelog/wiki

### Installing Prerequisite Libraries:
```
	go get github.com/streadway/amqp
    go get github.com/cihub/seelog

```

## Installing 
Due to the fact that 'go get' doesn't really understand gitlab based repositories 
that don't, in fact, begin with "gitlab." as part of the address, we have to 
download these libraries and install them locally:

```
    git clone https://git.synapse-wireless.com/hcap/rmqconsumer.git
    cd rmqconsumer
    go install
    
```
 
## Usage
The package defines one data structure (Consumer), one function type (ConsumerHandler), and
two exported functions (NewConsumer, Shutdown)

```
  import (
    rp "synapse-wireless.com/rmqpublisher"
    rc "synapse-wireless.com/rmqconsumer"
  )
```


### Data Types

#### Consumer 

```
    // The Consumer structure holds our connection to Rabbit
    type Consumer struct {
        conn    *amqp.Connection
        channel *amqp.Channel
        tag     string
        done    chan error
    }
```

#### ConsumerHandler

```
    // The function signature of our handler
    type ConsumerHandler func (<-chan amqp.Delivery, chan error)
```

### Functions

#### NewConsumer
Takes connection and Queue/Exchange parameters, establishers a connection to 
the specified RabbitMQ server, starts the provided Handler as a go routine and
then returns an initialized Consumer structure.

 * Note: If the sslcfg parameter is not nil, then rmqconsumer attempts to make an SSL connection to the server.

```
    func NewConsumer(amqpURI, exchange, exchangeType, queueName, key, ctag string, ch ConsumerHandler, sslcfg *tls.Config) (*Consumer, error)
```

#### Shutdown
Invoked via the returned Consumer struct, to shutdown the rabbit connection  and the handler.

```
    func (c *Consumer) Shutdown() error
```

### Sample Handler
This is the consumer handler from rabbitrelaygo
```
    func consumer_handler(deliveries <-chan amqp.Delivery, done chan error) {
        log.Trace("Started")
        for d := range deliveries {
            log.Trace("Processing deliveries")
            d.Ack(false)  // No matter what, we ack the message.
            // Publish the messages to the slave server(s)
            for _, p := range p_list {
                // Add the message to this publishers Queue
                p.InMsg <- d.Body 
                log.Trace("Message Sent")
            }
        }
        log.Trace("Deliveries channel closed.")
        done <- nil
    }
```

## Logging
Logging is handled via the seelog framework and configured by ./seelog.xml

 * Set the 'minlevel' to your minimum logging level.  One of:
  * "trace"
  * "debug"
  * "info"
  * "warn"
  * "error"
  * "critical"


Sample Logging Configuration file

```

<seelog minlevel="trace" maxlevel="critical">
    <outputs>
        <rollingfile type="size" filename="./rabbitrelay.log" maxsize="1000000" maxrolls="50" />
    </outputs>
</seelog>

```
