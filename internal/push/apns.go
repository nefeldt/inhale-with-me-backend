package push

import (
	"fmt"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// APNs is a token-authenticated (.p8) Apple Push Notification service client.
type APNs struct {
	client *apns2.Client
	topic  string // the app bundle id
}

// NewAPNs builds an APNs notifier from a .p8 signing key. production selects the
// production APNs host (used for TestFlight/App Store builds).
func NewAPNs(keyPEM []byte, keyID, teamID, bundleID string, production bool) (*APNs, error) {
	authKey, err := token.AuthKeyFromBytes(keyPEM)
	if err != nil {
		return nil, fmt.Errorf("parse APNs key: %w", err)
	}
	tok := &token.Token{AuthKey: authKey, KeyID: keyID, TeamID: teamID}
	client := apns2.NewTokenClient(tok)
	if production {
		client = client.Production()
	} else {
		client = client.Development()
	}
	return &APNs{client: client, topic: bundleID}, nil
}

// Send pushes an alert to each token, returning tokens APNs says are invalid.
func (a *APNs) Send(tokens []string, title, body string, data map[string]string) []string {
	var invalid []string
	for _, tk := range tokens {
		p := payload.NewPayload().AlertTitle(title).AlertBody(body).Sound("default")
		for k, v := range data {
			p = p.Custom(k, v)
		}
		res, err := a.client.Push(&apns2.Notification{DeviceToken: tk, Topic: a.topic, Payload: p})
		if err != nil {
			continue // transient network error — leave the token in place
		}
		switch res.Reason {
		case apns2.ReasonBadDeviceToken, apns2.ReasonUnregistered, apns2.ReasonDeviceTokenNotForTopic:
			invalid = append(invalid, tk)
		}
	}
	return invalid
}
