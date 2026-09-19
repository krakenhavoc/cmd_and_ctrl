package lobby

// botstats.go is #505 part 3: GET /games/{id}/bot/stats, an
// admin-only readout of every bot seat's own instrumentation — the
// half of #89's latency checklist that a human could not previously
// reach. It reuses exactly the metric shape #505 part 1 and part 2
// already built (aiseat.Stats, with its P50/P99/P999/Max latency
// percentiles, and aiseat.PolicyStats for the seats that have a model
// funnel underneath them) rather than inventing a second one.
//
// # Why this is its own file
//
// server/internal/lobby/http.go, where every other route is
// registered inline in Handler's mux.Handle block, and lobby.go,
// where BotHost is declared, are both large enough that a full
// re-emit of either is the kind of diff this repo's tooling treats
// as risky (a dropped line corrupts the file silently, and there is
// no cheap way to diff a whole-file rewrite against itself). Both are
// also under heavy concurrent churn from other work on this repo
// right now. So: a new file, a narrow interface that does not widen
// BotHost, and a handler that is fully wired and testable on its own
// — see BotStatsHandler's doc for the one line of registration it
// still needs once http.go (or main.go's own mux, which already
// composes routes this way — see its mux.Handle("/cards/", ...)
// beside mux.Handle("/", lobby.Handler(...))) can be safely touched.
//
// # Why a new interface rather than widening BotHost
//
// BotHost (lobby.go) is the seam #89's original bot-seat work drew
// between this package and aiseat: Tiers, StartBots, StopBots, and
// nothing a bot seat DOES beyond that. aiseat.Manager already
// exposes Runners(gameID) — "(tests, admin tools)", per its own doc
// comment — so BotStats below asks for exactly that, structurally,
// without *aiseat.Manager needing to change and without BotHost
// growing a method every caller that only starts and stops bots now
// has to satisfy. *aiseat.Manager keeps importing nothing of this
// package (see manager.go: aiseat stays the lower layer), so the
// dependency does not reverse.

import (
	"net/http"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
)

// BotStats is a narrow extension of BotHost: a bot host that can also
// list its live runners, for this admin-only readout. Satisfied by
// *aiseat.Manager without any change there — see the file comment.
//
// A BotHost that does NOT implement it (a test double built against
// the narrower interface, for instance) makes the endpoint answer
// 503 rather than panicking; see botStats below.
type BotStats interface {
	Runners(gameID uuid.UUID) []*aiseat.Runner
}

// BotStatsRoute is GET /games/{id}/bot/stats's method-and-pattern, in
// the exact form net/http's ServeMux (and every other route in
// Handler) expects. Exported so the one line that mounts it can live
// wherever this repo's size ceiling currently allows — see
// BotStatsHandler.
const BotStatsRoute = "GET /games/{id}/bot/stats"

// BotStatsHandler is GET /games/{id}/bot/stats, admin-only, wired
// exactly like every other admin route in Handler
// (auth.Middleware(c.Auth, auth.RoleAdmin) over the same handlerFunc
// adapter). It is a standalone http.Handler rather than an inline
// mux.Handle call in Handler itself so it can be mounted without
// editing http.go — see the file comment for why that matters right
// now. Once http.go can safely take the one-line addition, the
// canonical home for this route is beside the other admin routes
// there:
//
//	mux.Handle(BotStatsRoute, BotStatsHandler(c))
//
// Until then, main.go's own top-level mux already composes routes
// this way (mux.Handle("/cards/", ...), mux.Handle("/catalog", ...)
// beside mux.Handle("/", lobby.Handler(cfg))) and can take it the
// same way, with the SAME Config main.go already builds for
// lobby.Handler:
//
//	mux.Handle(lobby.BotStatsRoute, lobby.BotStatsHandler(cfg))
func BotStatsHandler(c Config) http.Handler {
	return auth.Middleware(c.Auth, auth.RoleAdmin)(handlerFunc(c, botStats))
}

// botSeatStats is one bot seat's line in the response: the runner's
// own counters (decisions, applied/rejected/fallback counts, the
// #505 part 1 latency percentiles, spend) and, for a seat with a
// model funnel underneath it, that funnel's own layer/escalation/
// token counters (#505 part 2). PolicyStats is omitted entirely for
// a `random` or `heuristic` seat, which is the true answer for a
// seat with no model.Policy in its chain.
type botSeatStats struct {
	Seat   uuid.UUID `json:"seat"`
	Policy string    `json:"policy"`
	// Stats is Runner.Stats(): this seat's own counters, including
	// Latency's P50/P99/P999/Max — #89's checklist item, finally
	// reachable from outside the process.
	Stats aiseat.Stats `json:"stats"`
	// PolicyStats is Runner.PolicyStats(), when this seat has one.
	PolicyStats *aiseat.PolicyStats `json:"policy_stats,omitempty"`
}

// botStatsResponse is one object per bot seat at the table, in the
// order Manager started their runners (seat order).
type botStatsResponse struct {
	Seats []botSeatStats `json:"seats"`
}

// botStats answers GET /games/{id}/bot/stats. 404 for an unknown
// game (the same check every other /games/{id} route makes, through
// c.Lobby.Get); 503 when this deployment has no bot host configured,
// or when the configured one does not implement BotStats — a
// deliberately narrower host, e.g. in a test, is a supported shape
// and not a server error. A game with no bot seats answers 200 with
// an empty list, which is the true state of that table.
func botStats(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	if _, err := c.Lobby.Get(id); err != nil {
		return err
	}
	if c.Bots == nil {
		return httpError(http.StatusServiceUnavailable, "bots are disabled on this deployment")
	}
	bs, ok := c.Bots.(BotStats)
	if !ok {
		return httpError(http.StatusServiceUnavailable, "bot stats are not available from this host")
	}
	runners := bs.Runners(id)
	out := make([]botSeatStats, 0, len(runners))
	for _, run := range runners {
		if run == nil {
			continue
		}
		row := botSeatStats{
			Seat:   run.Seat(),
			Policy: run.PolicyName(),
			Stats:  run.Stats(),
		}
		if ps, ok := run.PolicyStats(); ok {
			row.PolicyStats = &ps
		}
		out = append(out, row)
	}
	return writeJSON(w, http.StatusOK, botStatsResponse{Seats: out})
}
