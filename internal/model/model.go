// Package model holds the GORM-persisted entities and the API projection types
// shared across the store and HTTP layers.
package model

import "time"

// User is a registered account. PasswordHash is never serialized to JSON.
type User struct {
	ID           string    `gorm:"primaryKey;size:36" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null" json:"email"`
	Username     string    `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"not null" json:"-"`
	DisplayName  *string   `json:"display_name"`
	Bio          *string   `json:"bio"`
	AvatarURL    *string   `json:"avatar_url"`
	Currency     string    `gorm:"not null;default:EUR" json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PublicUser is the projection of a User shown to other users. It omits the
// email and never exposes the password hash. FriendStatus is filled relative to
// the requesting viewer when known.
type PublicUser struct {
	ID           string  `json:"id"`
	Username     string  `json:"username"`
	DisplayName  *string `json:"display_name"`
	AvatarURL    *string `json:"avatar_url"`
	FriendStatus string  `json:"friend_status,omitempty"`
}

// Public returns the PublicUser projection of u (without a friend status).
func (u User) Public() PublicUser {
	return PublicUser{
		ID:          u.ID,
		Username:    u.Username,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarURL,
	}
}

// SmokeSession is a single logged smoke.
type SmokeSession struct {
	ID         string      `gorm:"primaryKey;size:36" json:"id"`
	UserID     string      `gorm:"index;not null" json:"user_id"`
	Type       SessionType `gorm:"not null" json:"type"`
	Quantity   float64     `gorm:"not null;default:1" json:"quantity"`
	Note       *string     `json:"note"`
	Mood       *string     `json:"mood"`
	Location   *string     `json:"location"`
	CostCents  *int64      `json:"cost_cents"`
	Visibility Visibility  `gorm:"not null;default:friends" json:"visibility"`
	OccurredAt time.Time   `gorm:"index" json:"occurred_at"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// TableName pins the table name (GORM would otherwise pluralize to "smoke_sessions" anyway).
func (SmokeSession) TableName() string { return "smoke_sessions" }

// Friendship is a single directed row carrying the whole friendship lifecycle.
// requester_id + addressee_id is unique; an accepted row means the two users are
// friends regardless of direction.
type Friendship struct {
	ID          string           `gorm:"primaryKey;size:36" json:"id"`
	RequesterID string           `gorm:"not null;index;uniqueIndex:idx_friend_pair" json:"requester_id"`
	AddresseeID string           `gorm:"not null;index;uniqueIndex:idx_friend_pair" json:"addressee_id"`
	// PairKey is the canonical (order-independent) key of the two user ids. Its
	// unique index guarantees at most one friendship row per unordered pair, so
	// concurrent mutual requests can't create both (A,B) and (B,A).
	PairKey   string           `gorm:"uniqueIndex" json:"-"`
	Status    FriendshipStatus `gorm:"not null;default:pending" json:"status"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// Reaction is a user's reaction (default "cheers") on a session. One row per
// (session, user, type).
type Reaction struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	SessionID string    `gorm:"not null;index;uniqueIndex:idx_reaction_unique" json:"session_id"`
	UserID    string    `gorm:"not null;uniqueIndex:idx_reaction_unique" json:"user_id"`
	Type      string    `gorm:"not null;default:cheers;uniqueIndex:idx_reaction_unique" json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

// CostSetting is a per-user unit cost (in cents) for a session type, used to
// estimate spend in stats.
type CostSetting struct {
	UserID        string      `gorm:"primaryKey;size:36" json:"-"`
	Type          SessionType `gorm:"primaryKey" json:"type"`
	UnitCostCents int64       `gorm:"not null" json:"unit_cost_cents"`
}

// Device is an APNs push registration for a user's device.
type Device struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	UserID    string    `gorm:"index;not null" json:"user_id"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	Platform  string    `gorm:"not null;default:ios" json:"platform"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Block records that BlockerID has blocked BlockedID (App Store UGC requirement:
// users can block abusive users). Blocked users are hidden from each other's
// feeds and cannot send friend requests.
type Block struct {
	ID        string    `gorm:"primaryKey;size:36" json:"id"`
	BlockerID string    `gorm:"not null;index;uniqueIndex:idx_block_pair" json:"blocker_id"`
	BlockedID string    `gorm:"not null;index;uniqueIndex:idx_block_pair" json:"blocked_id"`
	CreatedAt time.Time `json:"created_at"`
}

// Report is a user's report of objectionable content (App Store UGC requirement:
// users can report content for review). SessionID is optional.
type Report struct {
	ID             string    `gorm:"primaryKey;size:36" json:"id"`
	ReporterID     string    `gorm:"not null;index" json:"reporter_id"`
	ReportedUserID string    `gorm:"not null;index" json:"reported_user_id"`
	SessionID      *string   `json:"session_id"`
	Reason         string    `json:"reason"`
	CreatedAt      time.Time `json:"created_at"`
}

// ---- Non-persisted API projection types ----

// ReactionSummary aggregates reaction counts for a session in the feed.
type ReactionSummary struct {
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
}

// FeedItem is one entry in a user's activity feed.
type FeedItem struct {
	Session         SmokeSession    `json:"session"`
	Author          PublicUser      `json:"author"`
	ReactionSummary ReactionSummary `json:"reaction_summary"`
	MyReactions     []string        `json:"my_reactions"`
}

// StatBucket is an aggregate over a time range.
type StatBucket struct {
	TotalCount         int                 `json:"total_count"`
	ByType             map[SessionType]int `json:"by_type"`
	TotalQuantity      float64             `json:"total_quantity"`
	EstimatedCostCents int64               `json:"estimated_cost_cents"`
	From               *time.Time          `json:"from"`
	To                 *time.Time          `json:"to"`
}

// StatsSummary is the dashboard payload returned by GET /stats/summary.
type StatsSummary struct {
	Timezone          string     `json:"timezone"`
	Currency          string     `json:"currency"`
	Today             StatBucket `json:"today"`
	Week              StatBucket `json:"week"`
	Month             StatBucket `json:"month"`
	AllTime           StatBucket `json:"all_time"`
	StreakDays        int        `json:"streak_days"`
	LongestStreakDays int        `json:"longest_streak_days"`
	DaysSinceLast     *int       `json:"days_since_last"`
}
