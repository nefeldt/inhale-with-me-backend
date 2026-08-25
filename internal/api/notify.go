package api

import "github.com/nfeldt/inhale-with-me/internal/model"

// displayName returns a friendly name for push copy.
func displayName(u *model.User) string {
	if u == nil {
		return "Someone"
	}
	if u.DisplayName != nil && *u.DisplayName != "" {
		return *u.DisplayName
	}
	return u.Username
}

// pushToUser sends a push to a single user's devices (fire-and-forget, no-op
// when push is unconfigured), pruning any tokens APNs reports as invalid.
func (a *API) pushToUser(userID, title, body string, data map[string]string) {
	go func() {
		devices, err := a.store.DevicesForUsers([]string{userID})
		if err != nil || len(devices) == 0 {
			return
		}
		tokens := make([]string, 0, len(devices))
		for _, d := range devices {
			tokens = append(tokens, d.Token)
		}
		for _, t := range a.notifier.Send(tokens, title, body, data) {
			_ = a.store.DeleteDeviceByToken(t)
		}
	}()
}
