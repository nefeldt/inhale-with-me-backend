package model

// SessionType is the kind of smoke that was logged.
type SessionType string

const (
	TypeCigarette SessionType = "cigarette"
	TypeJoint     SessionType = "joint"
	TypeVape      SessionType = "vape"
	TypeCigar     SessionType = "cigar"
	TypePipe      SessionType = "pipe"
	TypeOther     SessionType = "other"
)

// SessionTypes lists every valid session type.
var SessionTypes = []SessionType{TypeCigarette, TypeJoint, TypeVape, TypeCigar, TypePipe, TypeOther}

// Valid reports whether t is a known session type.
func (t SessionType) Valid() bool {
	for _, v := range SessionTypes {
		if v == t {
			return true
		}
	}
	return false
}

// Visibility controls who may see a session.
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityFriends Visibility = "friends"
	VisibilityPrivate Visibility = "private"
)

// Visibilities lists every valid visibility.
var Visibilities = []Visibility{VisibilityPublic, VisibilityFriends, VisibilityPrivate}

// Valid reports whether v is a known visibility.
func (v Visibility) Valid() bool {
	for _, x := range Visibilities {
		if x == v {
			return true
		}
	}
	return false
}

// Mood is an optional emotional tag on a session.
type Mood string

const (
	MoodGreat    Mood = "great"
	MoodGood     Mood = "good"
	MoodNeutral  Mood = "neutral"
	MoodStressed Mood = "stressed"
	MoodBad      Mood = "bad"
)

// Moods lists every valid mood.
var Moods = []Mood{MoodGreat, MoodGood, MoodNeutral, MoodStressed, MoodBad}

// Valid reports whether m is a known mood.
func (m Mood) Valid() bool {
	for _, x := range Moods {
		if x == m {
			return true
		}
	}
	return false
}

// FriendshipStatus is the lifecycle state of a friendship row.
type FriendshipStatus string

const (
	FriendshipPending  FriendshipStatus = "pending"
	FriendshipAccepted FriendshipStatus = "accepted"
	FriendshipDeclined FriendshipStatus = "declined"
)

// Friend-status values reported on a PublicUser relative to the viewer.
const (
	FriendStatusNone     = "none"
	FriendStatusIncoming = "incoming" // the other user sent the viewer a request
	FriendStatusOutgoing = "outgoing" // the viewer sent the other user a request
	FriendStatusFriends  = "friends"
)

// DefaultReactionType is used when a reaction is created without a type.
const DefaultReactionType = "cheers"
