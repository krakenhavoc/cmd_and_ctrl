package users

import (
	"context"
	"fmt"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// The database half of the metrics growth collector (ADR 0123
// amendment, #3557, metrics.GrowthSource). Read once every
// metrics.GrowthCacheTTL at most. Everything it returns is a count: the
// per-account rows it reads are folded in memory and never leave this
// function.

const growthDay = 24 * time.Hour

// Growth reads the growth numbers as of now. "Played" is what
// LastPlayed means: seated at a started table that ended, or was
// archived, at that time.
func (s *SQLStore) Growth(ctx context.Context, now time.Time) (metrics.GrowthReading, error) {
	out := metrics.GrowthReading{
		Windows: map[string]metrics.GrowthWindow{},
		Cohorts: map[string]metrics.GrowthCohort{},
	}
	t7, t14 := now.Add(-7*growthDay), now.Add(-14*growthDay)

	// One row per account that ever finished a game: its first and last
	// end, and whether it played in the week before the last one.
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.user_id,
		        MIN(COALESCE(g.ended_at, g.archived_at)),
		        MAX(COALESCE(g.ended_at, g.archived_at)),
		        MAX(COALESCE(g.ended_at, g.archived_at) >= ? AND COALESCE(g.ended_at, g.archived_at) < ?)
		   FROM seats s JOIN games g ON g.id = s.game_id
		  WHERE s.user_id IS NOT NULL
		    AND g.started_at IS NOT NULL
		    AND COALESCE(g.ended_at, g.archived_at) IS NOT NULL
		  GROUP BY s.user_id`,
		t14.UnixMilli(), t7.UnixMilli())
	if err != nil {
		return out, fmt.Errorf("users: growth players: %w", err)
	}
	defer func() { _ = rows.Close() }()
	lastEnd := map[string]int64{}
	for rows.Next() {
		var (
			id          string
			first, last int64
			prevWeek    bool
		)
		if err := rows.Scan(&id, &first, &last, &prevWeek); err != nil {
			return out, fmt.Errorf("users: growth players: %w", err)
		}
		lastEnd[id] = last
		for _, w := range windowSpans() {
			since := now.Add(-w.span).UnixMilli()
			if last < since {
				continue
			}
			x := out.Windows[w.label]
			if first >= since {
				x.NewPlayers++
			} else {
				x.ReturningPlayers++
			}
			out.Windows[w.label] = x
		}
		if prevWeek {
			out.RetentionPrevious++
			if last >= t7.UnixMilli() {
				out.RetentionRetained++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return out, fmt.Errorf("users: growth players: %w", err)
	}

	// The signup cohorts: accounts created in each band, and how many of
	// them played in the last 7 days.
	oldest := 0
	for _, label := range metrics.SignupCohortLabels() {
		w, _ := metrics.SignupCohortWeeks(label)
		oldest = max(oldest, w)
	}
	urows, err := s.db.QueryContext(ctx,
		`SELECT id, created_at FROM users WHERE created_at >= ? AND created_at < ?`,
		now.Add(-time.Duration(oldest+1)*7*growthDay).UnixMilli(), t7.UnixMilli())
	if err != nil {
		return out, fmt.Errorf("users: growth cohorts: %w", err)
	}
	defer func() { _ = urows.Close() }()
	for urows.Next() {
		var (
			id      string
			created int64
		)
		if err := urows.Scan(&id, &created); err != nil {
			return out, fmt.Errorf("users: growth cohorts: %w", err)
		}
		for _, label := range metrics.SignupCohortLabels() {
			weeks, _ := metrics.SignupCohortWeeks(label)
			hi := now.Add(-time.Duration(weeks) * 7 * growthDay).UnixMilli()
			lo := now.Add(-time.Duration(weeks+1) * 7 * growthDay).UnixMilli()
			if created < lo || created >= hi {
				continue
			}
			c := out.Cohorts[label]
			c.Accounts++
			if lastEnd[id] >= t7.UnixMilli() {
				c.Active++
			}
			out.Cohorts[label] = c
		}
	}
	if err := urows.Err(); err != nil {
		return out, fmt.Errorf("users: growth cohorts: %w", err)
	}

	// Engagement: the finished games (those with statistics) per window.
	for _, w := range windowSpans() {
		since := now.Add(-w.span).UnixMilli()
		x := out.Windows[w.label]
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(*), COALESCE(SUM(gs.turns), 0), COALESCE(SUM(gs.rounds), 0),
			        COALESCE(SUM((SELECT COUNT(*) FROM seats s WHERE s.game_id = gs.game_id)), 0)
			   FROM game_stats gs JOIN games g ON g.id = gs.game_id
			  WHERE g.ended_at >= ?`, since).
			Scan(&x.GamesFinished, &x.Turns, &x.Rounds, &x.Seats); err != nil {
			return out, fmt.Errorf("users: growth games: %w", err)
		}
		if err := s.db.QueryRowContext(ctx,
			`SELECT COUNT(DISTINCT s.user_id)
			   FROM seats s JOIN game_stats gs ON gs.game_id = s.game_id
			   JOIN games g ON g.id = s.game_id
			  WHERE g.ended_at >= ?`, since).
			Scan(&x.FinishedPlayers); err != nil {
			return out, fmt.Errorf("users: growth players per game: %w", err)
		}
		out.Windows[w.label] = x
	}
	return out, nil
}

// windowSpans are the users collector's windows (1d, 7d, 30d), so the
// growth gauges count over the same spans.
func windowSpans() []struct {
	label string
	span  time.Duration
} {
	var out []struct {
		label string
		span  time.Duration
	}
	for _, label := range []string{"1d", "7d", "30d"} {
		if span, ok := metrics.UsersPlayedWindow(label); ok {
			out = append(out, struct {
				label string
				span  time.Duration
			}{label, span})
		}
	}
	return out
}
