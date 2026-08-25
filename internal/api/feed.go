package api

import (
	"net/http"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

func (a *API) handleFeed(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)

	limit := queryInt(r, "limit", 20)
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	before := queryTime(r, "before")

	friendIDs, err := a.store.FriendIDs(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	sessions, err := a.store.FeedSessions(friendIDs, before, limit)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	sessionIDs := make([]string, 0, len(sessions))
	authorIDs := make(map[string]struct{})
	for _, s := range sessions {
		sessionIDs = append(sessionIDs, s.ID)
		authorIDs[s.UserID] = struct{}{}
	}

	summaries, mine, err := a.store.SummariesForSessions(sessionIDs, me)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	authors := make(map[string]model.PublicUser, len(authorIDs))
	for id := range authorIDs {
		if u, err := a.store.GetUserByID(id); err == nil {
			authors[id] = u.Public()
		}
	}

	items := make([]model.FeedItem, 0, len(sessions))
	for _, s := range sessions {
		sum := summaries[s.ID]
		if sum.Counts == nil {
			sum.Counts = map[string]int{}
		}
		reactions := mine[s.ID]
		if reactions == nil {
			reactions = []string{}
		}
		items = append(items, model.FeedItem{
			Session:         s,
			Author:          authors[s.UserID],
			ReactionSummary: sum,
			MyReactions:     reactions,
		})
	}

	resp := map[string]any{"items": items, "next_before": nil}
	if len(sessions) == limit {
		resp["next_before"] = sessions[len(sessions)-1].OccurredAt
	}
	writeJSON(w, http.StatusOK, resp)
}
