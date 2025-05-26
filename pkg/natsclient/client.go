package natsclient

import (
	"context"
	"time"

	"github.com/charmbracelet/log"
	"github.com/nats-io/nats.go"
)
 
 // Connect establishes a connection to a NATS server.
 // It takes the NATS server URL and a variadic slice of nats.Option as arguments.
 // This allows for flexible configuration (credentials, connection name, etc.).
 func Connect(natsURL string, opts ...nats.Option) (*nats.Conn, error) {
 	// Apply default options first, then allow user-provided options to override or add.
 	defaultOpts := []nats.Option{
 		nats.Name("Gomem NATS Client"), // You might want to make the client name configurable
 		nats.Timeout(10 * time.Second),
 		nats.RetryOnFailedConnect(true),
 		nats.MaxReconnects(10), // Increased max reconnects
 		nats.ReconnectWait(3 * time.Second),
 		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
 			log.Printf("NATS client disconnected. Last error: %v", err)
 		}),
 		nats.ReconnectHandler(func(nc *nats.Conn) {
 			log.Printf("NATS client reconnected to %s", nc.ConnectedUrl())
 		}),
 		nats.ClosedHandler(func(nc *nats.Conn) {
 			log.Info("NATS client connection closed.")
 		}),
 	}
 
 	finalOpts := append(defaultOpts, opts...)
 
 	// Connect to NATS
 	nc, err := nats.Connect(natsURL, finalOpts...)
 	if err != nil {
 		log.Printf("Error connecting to NATS at %s: %v", natsURL, err)
 		return nil, err
 	}
 
 	log.Printf("Successfully connected to NATS at %s", nc.ConnectedUrl())
 	return nc, nil
 }
 
 // Publish sends a message to the given subject.
 func Publish(nc *nats.Conn, subject string, data []byte) error {
 	if nc == nil {
 		log.Error("NATS connection is not established.")
 		return nats.ErrConnectionClosed
 	}
 	err := nc.Publish(subject, data)
 	if err != nil {
 		log.Errorf("Error publishing message to subject %s: %v", subject, err)
 		return err
 	}
 	log.Infof("Message published to subject %s", subject)
 	return nil
 }
 
 // Subscribe creates a subscription to the given subject.
 // The provided handler function will be called for each message received.
 func Subscribe(nc *nats.Conn, subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
 	if nc == nil {
 		log.Error("NATS connection is not established.")
 		return nil, nats.ErrConnectionClosed
 	}
 	sub, err := nc.Subscribe(subject, handler)
 	if err != nil {
 		log.Errorf("Error subscribing to subject %s: %v", subject, err)
 		return nil, err
 	}
 	log.Infof("Subscribed to subject %s", subject)
 	return sub, nil
 }
 
 // Request sends a request message and waits for a response.
 // It uses a context for timeout and cancellation.
 func Request(nc *nats.Conn, subject string, data []byte, timeout time.Duration) (*nats.Msg, error) {
 	if nc == nil {
 		log.Error("NATS connection is not established.")
 		return nil, nats.ErrConnectionClosed
 	}
 	ctx, cancel := context.WithTimeout(context.Background(), timeout)
 	defer cancel()
 
 	msg, err := nc.RequestWithContext(ctx, subject, data)
 	if err != nil {
 		log.Errorf("Error making request to subject %s: %v", subject, err)
 		return nil, err
 	}
 	log.Infof("Received response from subject %s", subject)
 	return msg, nil
 }