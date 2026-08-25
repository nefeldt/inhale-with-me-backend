package api

import (
	"math"
	"net/http"
	"sort"
	"time"

	"github.com/nfeldt/inhale-with-me/internal/model"
)

// handleStats returns the dashboard summary (today/week/month/all-time buckets
// plus streaks) computed in the caller's timezone.
func (a *API) handleStats(w http.ResponseWriter, r *http.Request) {
	me := currentUserID(r)

	tz := r.URL.Query().Get("tz")
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
		tz = "UTC"
	}

	u, err := a.store.GetUserByID(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	costMap, err := a.store.CostMap(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	sessions, err := a.store.AllSessionsByUser(me)
	if err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, computeStats(sessions, costMap, loc, tz, u.Currency))
}

func computeStats(sessions []model.SmokeSession, costMap map[model.SessionType]int64, loc *time.Location, tz, currency string) model.StatsSummary {
	now := time.Now().In(loc)
	todayStart := dayStart(now)
	weekStart := mondayStart(now)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

	current, longest, daysSinceLast := computeStreaks(sessions, loc)

	return model.StatsSummary{
		Timezone:          tz,
		Currency:          currency,
		Today:             bucket(sessions, &todayStart, costMap),
		Week:              bucket(sessions, &weekStart, costMap),
		Month:             bucket(sessions, &monthStart, costMap),
		AllTime:           bucket(sessions, nil, costMap),
		StreakDays:        current,
		LongestStreakDays: longest,
		DaysSinceLast:     daysSinceLast,
	}
}

// bucket aggregates sessions occurring at/after `from` (nil = all time).
func bucket(sessions []model.SmokeSession, from *time.Time, costMap map[model.SessionType]int64) model.StatBucket {
	b := model.StatBucket{ByType: make(map[model.SessionType]int, len(model.SessionTypes))}
	// Serialize the boundary as RFC3339 UTC per the API contract (it is still the
	// same instant as the local midnight used for filtering below).
	if from != nil {
		u := from.UTC()
		b.From = &u
	}
	for _, t := range model.SessionTypes {
		b.ByType[t] = 0
	}
	for _, s := range sessions {
		if from != nil && s.OccurredAt.Before(*from) {
			continue
		}
		b.TotalCount++
		b.ByType[s.Type]++
		b.TotalQuantity += s.Quantity
		b.EstimatedCostCents += sessionCost(s, costMap)
	}
	return b
}

// sessionCost returns the estimated cost of a session in cents: an explicit
// per-session cost if set, otherwise quantity * the user's per-type unit cost.
func sessionCost(s model.SmokeSession, costMap map[model.SessionType]int64) int64 {
	if s.CostCents != nil {
		return *s.CostCents
	}
	if unit, ok := costMap[s.Type]; ok {
		return int64(math.Round(float64(unit) * s.Quantity))
	}
	return 0
}

// computeStreaks returns the current daily streak, the longest ever streak and
// the whole days since the last session (nil if there are no sessions).
func computeStreaks(sessions []model.SmokeSession, loc *time.Location) (current, longest int, daysSinceLast *int) {
	if len(sessions) == 0 {
		return 0, 0, nil
	}

	dateSet := make(map[string]bool)
	var last time.Time
	for _, s := range sessions {
		local := s.OccurredAt.In(loc)
		dateSet[local.Format("2006-01-02")] = true
		if s.OccurredAt.After(last) {
			last = s.OccurredAt
		}
	}

	// Days since last session, counted in whole calendar days (DST-safe).
	ds := civilDaysBetween(last.In(loc), time.Now().In(loc))
	if ds < 0 {
		ds = 0
	}

	// Current streak: consecutive days ending today (or yesterday if not logged today yet).
	day := dayStart(time.Now().In(loc))
	if !dateSet[day.Format("2006-01-02")] {
		day = day.AddDate(0, 0, -1)
	}
	for dateSet[day.Format("2006-01-02")] {
		current++
		day = day.AddDate(0, 0, -1)
	}

	// Longest streak: scan the sorted set of distinct days.
	days := make([]time.Time, 0, len(dateSet))
	for k := range dateSet {
		if d, err := time.ParseInLocation("2006-01-02", k, loc); err == nil {
			days = append(days, d)
		}
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Before(days[j]) })
	longest, run := 1, 1
	for i := 1; i < len(days); i++ {
		if days[i].Equal(days[i-1].AddDate(0, 0, 1)) {
			run++
		} else {
			run = 1
		}
		if run > longest {
			longest = run
		}
	}

	return current, longest, &ds
}

func dayStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// civilDaysBetween returns the number of whole calendar days from -> to, using
// the local calendar dates. Anchoring both civil dates to UTC midnight avoids
// the DST bug of dividing an elapsed Duration by 24h (a spring-forward day is
// only 23h of wall-clock time).
func civilDaysBetween(from, to time.Time) int {
	a := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)
	return int(b.Sub(a).Hours() / 24)
}

// mondayStart returns 00:00 on the Monday of t's week.
func mondayStart(t time.Time) time.Time {
	d := dayStart(t)
	offset := (int(d.Weekday()) + 6) % 7 // Monday=0 ... Sunday=6
	return d.AddDate(0, 0, -offset)
}
