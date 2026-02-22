package models

// User user model.
type User struct {
	ID           int
	Username     string
	PasswordHash string
}

// PushSubscription subscriptions model
type PushSubscription struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
}
