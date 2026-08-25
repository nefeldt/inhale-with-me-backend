// Package push delivers notifications to user devices.
package push

// Notifier delivers push notifications. Send returns the device tokens that the
// provider reported as permanently invalid, so the caller can prune them.
type Notifier interface {
	Send(tokens []string, title, body string, data map[string]string) []string
}

// Noop is a Notifier that does nothing — used when APNs is not configured, so
// the rest of the app runs unchanged without push credentials.
type Noop struct{}

// Send does nothing and reports no invalid tokens.
func (Noop) Send([]string, string, string, map[string]string) []string { return nil }
