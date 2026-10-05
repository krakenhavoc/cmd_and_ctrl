package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// The games and players event metrics (ADR 0123 §3, "Games and
// players"). The lobby records them; the state gauges beside them
// (cmdctrl_games, cmdctrl_seats, …) are the collectors in tables.go
// and users.go.

// Outcomes of cmdctrl_games_ended_total.
const (
	// OutcomeWin is an ended game with a winner (games.outcome 'win').
	OutcomeWin = "win"
	// OutcomeDraw is an ended game nobody won (games.outcome 'draw').
	OutcomeDraw = "draw"
	// OutcomeClosed is a table that stopped without the engine ending
	// it with a result: an operator archived or deleted it while it
	// was running, or the engine ended it with no outcome recorded.
	// In migration 0007's terms, an ended game whose outcome is NULL.
	OutcomeClosed = "closed"
)

var outcomeLabels = []string{OutcomeWin, OutcomeDraw, OutcomeClosed}

// maxSeatsLabel bounds the seats label of cmdctrl_games_started_total:
// the ADR's closed set is 1…8. game.MaxPlayers is 4 today; a count
// outside the set is clamped into it rather than minting a new value.
const maxSeatsLabel = 8

var seatsLabels = func() []string {
	out := make([]string, 0, maxSeatsLabel)
	for i := 1; i <= maxSeatsLabel; i++ {
		out = append(out, strconv.Itoa(i))
	}
	return out
}()

var (
	gamesCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_games_created_total",
		Help: "Tables created through POST /games.",
	})

	gamesStarted = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_games_started_total",
		Help: "Tables that left the lobby, by how many seats they started with. Practice tables are not counted.",
	}, []string{"seats"})

	gamesEnded = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "cmdctrl_games_ended_total",
		Help: "Tables that ended, by outcome: win, draw, or closed (archived or deleted while running, or ended with no result). Practice tables are not counted.",
	}, []string{"outcome"})

	gameDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name: "cmdctrl_game_duration_seconds",
		Help: "Time from a table's start to its end, observed when it ends. A table with no recorded start is not observed.",
		// Five minutes to a day: a concession on turn two to a table
		// left running overnight.
		Buckets: []float64{
			5 * 60, 10 * 60, 15 * 60, 20 * 60, 30 * 60, 45 * 60,
			60 * 60, 90 * 60, 2 * 3600, 3 * 3600, 4 * 3600, 6 * 3600, 12 * 3600, 24 * 3600,
		},
	})

	usersCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "cmdctrl_users_created_total",
		Help: "New accounts: first Discord sign-ins, each a new users row.",
	})
)

func init() {
	// Every series exists from boot at zero, so increase() sees the
	// first event rather than a series appearing at one.
	for _, s := range seatsLabels {
		gamesStarted.WithLabelValues(s)
	}
	for _, o := range outcomeLabels {
		gamesEnded.WithLabelValues(o)
	}
}

// GameCreated counts one successful POST /games.
func GameCreated() { gamesCreated.Inc() }

// GameStarted counts a table leaving the lobby with seats seats.
func GameStarted(seats int) {
	seats = max(1, min(seats, maxSeatsLabel))
	gamesStarted.WithLabelValues(strconv.Itoa(seats)).Inc()
}

// EndOutcome maps the engine's outcome kind (game.OutcomeKind: "win",
// "draw", or "" for none) onto the outcome label. Anything but a win
// or a draw is closed, as a NULL games.outcome is.
func EndOutcome(kind string) string {
	switch kind {
	case OutcomeWin, OutcomeDraw:
		return kind
	}
	return OutcomeClosed
}

// GameEnded counts one table's end with outcome (OutcomeWin,
// OutcomeDraw or OutcomeClosed; anything else counts as closed) and
// observes its length. A negative length means the start is not
// known (a table imported with no started_at), and observes nothing.
func GameEnded(outcome string, length time.Duration) {
	gamesEnded.WithLabelValues(EndOutcome(outcome)).Inc()
	if length >= 0 {
		gameDuration.Observe(length.Seconds())
	}
}

// UserCreated counts a new users row.
func UserCreated() { usersCreated.Inc() }
