package subscriber

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func ConnectWithBackoff(url string, maxAttempts int) (*nats.Conn, error) {
	delay := time.Second
	for i := 1; i <= maxAttempts; i++ {
		nc, err := nats.Connect(url)
		if err == nil {
			slog.Info("connected to NATS", "url", url)
			return nc, nil
		}
		slog.Warn("NATS connect attempt failed", "attempt", i, "max_attempts", maxAttempts, "error", err, "retry_in", delay.String())
		time.Sleep(delay)
		delay *= 2
	}
	return nil, fmt.Errorf("could not connect to NATS at %s after %d attempts", url, maxAttempts)
}

func Subscribe(nc *nats.Conn, subjects []string) {
	for _, subj := range subjects {
		s := subj
		_, err := nc.Subscribe(s, func(msg *nats.Msg) {
			handle(s, msg.Data)
		})
		if err != nil {
			slog.Error("failed to subscribe", "subject", s, "error", err)
		} else {
			slog.Info("subscribed", "subject", s)
		}
	}
}

func handle(subject string, data []byte) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		slog.Error("failed to unmarshal event", "subject", subject, "error", err)
		return
	}

	out := map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"subject": subject,
		"event":   payload,
	}
	b, _ := json.Marshal(out)
	os.Stdout.Write(b)
	os.Stdout.Write([]byte("\n"))
}
