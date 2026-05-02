package subscriber

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

func ConnectWithBackoff(url string, maxAttempts int) (*nats.Conn, error) {
	delay := time.Second
	for i := 1; i <= maxAttempts; i++ {
		nc, err := nats.Connect(url)
		if err == nil {
			log.Printf("Connected to NATS at %s", url)
			return nc, nil
		}
		log.Printf("NATS connect attempt %d/%d failed: %v — retrying in %s", i, maxAttempts, err, delay)
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
			log.Printf("ERROR: failed to subscribe to %s: %v", s, err)
		} else {
			log.Printf("Subscribed to %s", s)
		}
	}
}

func handle(subject string, data []byte) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		fmt.Fprintf(os.Stderr, `{"level":"error","subject":%q,"error":%q}`+"\n", subject, err.Error())
		return
	}

	out := map[string]any{
		"time":    time.Now().UTC().Format(time.RFC3339),
		"subject": subject,
		"event":   payload,
	}
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}
