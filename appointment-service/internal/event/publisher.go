package event

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

type Publisher interface {
	Publish(ctx context.Context, subject string, payload any) error
}

type natsPublisher struct {
	nc *nats.Conn
}

type noopPublisher struct{}

func NewPublisher(url string) Publisher {
	nc, err := nats.Connect(url,
		nats.ReconnectWait(2*nats.DefaultReconnectWait),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		log.Printf("WARN: NATS unavailable at %s: %v — events will be dropped", url, err)
		return &noopPublisher{}
	}
	return &natsPublisher{nc: nc}
}

func (p *natsPublisher) Publish(_ context.Context, subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal subject=%s: %w", subject, err)
	}
	if err := p.nc.Publish(subject, data); err != nil {
		return fmt.Errorf("nats publish subject=%s: %w", subject, err)
	}
	return nil
}

func (p *noopPublisher) Publish(_ context.Context, subject string, _ any) error {
	log.Printf("WARN: noop publisher — event dropped subject=%s", subject)
	return nil
}
