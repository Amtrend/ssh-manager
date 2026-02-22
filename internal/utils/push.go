package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ssh_manager/internal/models"
	"strconv"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
)

// SendNotification sending notifications to clients.
func SendNotification(sub models.PushSubscription, title, body string) error {
	// Telling the browser what exactly to show.
	payload, err := json.Marshal(map[string]string{
		"title": title,
		"body":  body,
	})

	if err != nil {
		LogErrorf("[PUSH] Payload marshal error: %v", err)
		return err
	}

	s := &webpush.Subscription{
		Endpoint: sub.Endpoint,
		Keys: webpush.Keys{
			P256dh: sub.P256dh,
			Auth:   sub.Auth,
		},
	}

	publicKey := strings.TrimSpace(GetEnv("VAPID_PUBLIC_KEY", ""))
	privateKey := strings.TrimSpace(GetEnv("VAPID_PRIVATE_KEY", ""))
	rawSub := strings.TrimSpace(GetEnv("VAPID_SUBSCRIBER_EMAIL", "admin@example.com"))
	subscriber := strings.Trim(rawSub, " \"\n\r\t")

	ttl, _ := strconv.Atoi(GetEnv("PUSH_TTL", "3600"))
	if ttl == 0 {
		ttl = 3600
	}

	subscriber = strings.TrimPrefix(subscriber, "mailto:")
	subscriber = "mailto:" + subscriber

	// Sending a request to the Push server (Google/Mozilla/Apple).
	resp, err := webpush.SendNotification(payload, s, &webpush.Options{
		Subscriber:      subscriber,
		VAPIDPublicKey:  publicKey,
		VAPIDPrivateKey: privateKey,
		TTL:             ttl,
		Urgency:         webpush.UrgencyHigh,
	})

	if err != nil {
		LogErrorf("[PUSH] Network or Library error: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		errDetail := fmt.Errorf("status: %d, body: %s", resp.StatusCode, string(respBody))

		LogErrorf("[PUSH] Server rejected request: %v", errDetail)

		return errDetail
	}
	return nil
}
