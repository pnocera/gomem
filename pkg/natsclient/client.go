package natsclient

import (
	"github.com/charmbracelet/log"
	"time"

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
