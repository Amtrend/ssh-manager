package utils

import (
	"encoding/json"
	"ssh_manager/internal/models"
	"strconv"

	"github.com/SherClockHolmes/webpush-go"
)

// SendNotification sending notifications to clients.
func SendNotification(sub models.PushSubscription, title, body string) error {
	// Telling the browser what exactly to show.
	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"body":  body,
	})

	s := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	publicKey := GetEnv("VAPID_PUBLIC_KEY", "")
	privateKey := GetEnv("VAPID_PRIVATE_KEY", "")
	subscriber := GetEnv("VAPID_SUBSCRIBER_EMAIL", "mailto:admin@example.com")

	ttl, _ := strconv.Atoi(GetEnv("PUSH_TTL", "3600"))
	if ttl == 0 {
		ttl = 3600
	}

	// Sending a request to the Push server (Google/Mozilla/Apple).
	resp, err := webpush.SendNotification(payload, s, &webpush.Options{
		Subscriber:      subscriber,
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
		TTL:             ttl,
	})

	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
