package lobby

// results_feed.go: a finished game's post in the Discord #game-results
// channel (ADR 0174). The host's choice (games.results), PUT
// /games/{id}/results, and the post itself, made once from the same
// moment and the same computed record as the game's statistics
// (gamestats.go, ADR 0165).
//
// Unlike the #lfg board there is no manager: the post is one call at one
// moment, so it runs on the stats goroutine, off l.mu, and a failure is
// logged and forgotten.

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// The ADR's numbers (§3, §5). Constants, not settings.
const (
	// resultsMinTurns is the turn a game must have reached before a
	// concession is worth posting. A game that ends by concession before
	// turn 3 is never posted (ADR 0174 §3).
	resultsMinTurns = 3
	// resultsMaxAttempts is how many transient failures in a row a post
	// tolerates before it is dropped: the #lfg manager's count.
	resultsMaxAttempts = 3
	// resultsCallTimeout bounds one Discord call.
	resultsCallTimeout = 10 * time.Second
	// resultsMaxRetryWait caps a 429's wait, and resultsBackoff is the
	// wait after a 5xx or a timeout.
	resultsMaxRetryWait = 30 * time.Second
	resultsBackoff      = 2 * time.Second
)

var (
	// ErrResultsOff: this server has no results channel (ADR 0174 §6).
	// 503 at the HTTP layer, naming what is missing.
	ErrResultsOff = errors.New("lobby: the results feed is off on this server")
	// ErrResultsUnavailable: this table cannot be opted in. It is the
	// tutorial's practice table, its game has ended, or it has been
	// archived. Turning the post off is always allowed.
	ErrResultsUnavailable = errors.New("lobby: only a table that has not ended can be opted in to the results feed")
)

// resultsAPI is the Discord half the feed uses: discord.Results, or a
// fake in tests.
type resultsAPI interface {
	Post(ctx context.Context, client *http.Client, p discord.ResultsPost) error
}

// ResultsFeed posts finished games. Construct it with NewResultsFeed and
// attach it with Lobby.SetResultsFeed.
type ResultsFeed struct {
	api    resultsAPI
	client *http.Client
	log    *slog.Logger
	// wait sleeps between attempts; a test replaces it.
	wait func(time.Duration)
	// minTurns is resultsMinTurns; a test lowers it to post a game it
	// can end without playing three turns.
	minTurns int
}

// NewResultsFeed builds a feed over the configured #game-results call.
func NewResultsFeed(r discord.Results, log *slog.Logger) *ResultsFeed {
	return newResultsFeed(r, &http.Client{Timeout: resultsCallTimeout}, log)
}

func newResultsFeed(api resultsAPI, client *http.Client, log *slog.Logger) *ResultsFeed {
	if log == nil {
		log = slog.Default()
	}
	return &ResultsFeed{api: api, client: client, log: log, wait: time.Sleep, minTurns: resultsMinTurns}
}

// post sends one post: up to resultsMaxAttempts tries while Discord is
// rate-limiting or not answering, and no more for any other error. A
// failure is logged and nothing else (the game is long over).
func (f *ResultsFeed) post(id uuid.UUID, p discord.ResultsPost) {
	for attempt := 1; ; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), resultsCallTimeout)
		err := f.api.Post(ctx, f.client, p)
		cancel()
		if err == nil {
			return
		}
		if !discord.Transient(err) || attempt >= resultsMaxAttempts {
			f.log.Warn("results feed: post failed", "game_id", id, "attempt", attempt, "err", err)
			return
		}
		wait := resultsBackoff
		var rl *discord.RateLimitError
		if errors.As(err, &rl) {
			wait = min(rl.RetryAfter, resultsMaxRetryWait)
		}
		f.wait(wait)
	}
}

// SetResultsFeed attaches the results feed, which turns it on. Called
// once at boot.
func (l *Lobby) SetResultsFeed(f *ResultsFeed) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.results = f
}

// ResultsEnabled reports whether this server has a results channel.
func (l *Lobby) ResultsEnabled() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.results != nil
}

// SetResults records the host's choice for a table (ADR 0174 §1). The
// caller has checked the host; this checks the table. Turning it on
// needs a table whose game has not ended and that is not archived or the
// practice table. Setting the value it already has is a no-op that still
// answers the meta.
func (l *Lobby) SetResults(id uuid.UUID, enabled bool) (GameMeta, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.results == nil {
		return GameMeta{}, ErrResultsOff
	}
	entry, ok := l.games[id]
	if !ok {
		return GameMeta{}, ErrGameNotFound
	}
	if entry.practice {
		return GameMeta{}, ErrResultsUnavailable
	}
	l.syncStateLocked(entry)
	if enabled && (entry.meta.State == string(game.StateEnded) || entry.meta.Archived()) {
		return GameMeta{}, ErrResultsUnavailable
	}
	if entry.meta.Results != enabled {
		entry.meta.Results = enabled
		l.persistGameLocked(entry)
	}
	return copyMeta(entry.meta), nil
}

// resultsJob is what the post needs from the table, captured under l.mu
// when the game ends.
type resultsJob struct {
	feed    *ResultsFeed
	gameID  uuid.UUID
	table   string
	outcome string
	// from is when the game started, or the table opened when no start is
	// on record; to is when it ended.
	from, to time.Time
}

// resultsJobLocked is the job for entry, nil when it is not to be
// posted: no feed, a practice table, or a table whose host did not opt
// in. Callers hold l.mu.
func (l *Lobby) resultsJobLocked(entry *gameEntry) *resultsJob {
	if l.results == nil || entry.practice || !entry.meta.Results {
		return nil
	}
	job := &resultsJob{
		feed: l.results, gameID: entry.meta.ID, table: entry.meta.Name,
		outcome: entry.outcome, from: entry.meta.CreatedAt, to: time.Now(),
	}
	if entry.startedAt != nil {
		job.from = *entry.startedAt
	}
	if entry.endedAt != nil {
		job.to = *entry.endedAt
	}
	return job
}

// run builds the post from the computed record and sends it, or says why
// it did not.
func (j *resultsJob) run(view protocol.GameView, rec gamestats.Game) {
	p, ok := buildResultsPost(j, view, rec)
	if !ok {
		return
	}
	j.feed.post(j.gameID, p)
}

// buildResultsPost is the post for a finished game, and false when the
// game is not to be posted (ADR 0174 §3): it has no outcome (it was
// closed, not played out), or it was conceded before resultsMinTurns.
func buildResultsPost(j *resultsJob, view protocol.GameView, rec gamestats.Game) (discord.ResultsPost, bool) {
	if j.outcome == "" || len(rec.Seats) == 0 {
		return discord.ResultsPost{}, false
	}
	if rec.Turns < j.feed.minTurns {
		for _, s := range rec.Seats {
			if s.Conceded {
				return discord.ResultsPost{}, false
			}
		}
	}
	players := make(map[int]protocol.PlayerView, len(view.Seats))
	for _, pv := range view.Seats {
		players[pv.Seat] = pv
	}
	seats := append([]gamestats.Seat(nil), rec.Seats...)
	// The winner first, then by place; a seat with no place goes last.
	sort.SliceStable(seats, func(a, b int) bool {
		pa, pb := placeOf(seats[a]), placeOf(seats[b])
		if pa != pb {
			return pa < pb
		}
		return seats[a].Seat < seats[b].Seat
	})
	post := discord.ResultsPost{Table: j.table, Turns: rec.Turns, Duration: j.to.Sub(j.from), EndedAt: j.to}
	for _, s := range seats {
		pv := players[s.Seat]
		rs := discord.ResultsSeat{
			Name:       pv.Name,
			Commanders: s.Commanders,
			Place:      shownPlace(s),
			Result:     string(s.Result),
			Conceded:   s.Conceded,
		}
		if s.EliminatedTurn != nil {
			rs.EliminatedTurn = *s.EliminatedTurn
		}
		switch {
		case pv.IsAgent:
			rs.Kind = "AI agent"
		case pv.IsBot:
			rs.Kind = "bot"
		}
		post.Seats = append(post.Seats, rs)
	}
	return post, true
}

// shownPlace is the seat's place, 0 when the game gave it none.
func shownPlace(s gamestats.Seat) int {
	if s.Place == nil {
		return 0
	}
	return *s.Place
}

func placeOf(s gamestats.Seat) int {
	if s.Place == nil {
		return 1 << 20
	}
	return *s.Place
}

// resultsRequest is the body of PUT /games/{id}/results. Enabled is a
// pointer so an empty body is a 400, not a silent "off".
type resultsRequest struct {
	Enabled *bool `json:"enabled"`
}

// putResults handles PUT /games/{id}/results: the host (or an admin)
// opts the table in to the results feed, or out of it (ADR 0174 §1). It
// answers the table's meta, like PUT /games/{id}/lfg.
func putResults(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return httpError(http.StatusInternalServerError, "missing principal")
	}
	var body resultsRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if body.Enabled == nil {
		return httpError(http.StatusBadRequest, "enabled is required")
	}
	if !c.Lobby.ResultsEnabled() {
		return errResultsOffHTTP
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !CanManageTable(p, meta, c.isAdmin(p)) {
		return ErrNotTableManager
	}
	meta, err = c.Lobby.SetResults(id, *body.Enabled)
	if err := resultsHTTPError(err); err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, redactMetaFor(p, c.isAdmin(p), id, meta))
}

// resultsHTTPError maps SetResults's refusals onto statuses.
func resultsHTTPError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ErrResultsOff):
		return errResultsOffHTTP
	case errors.Is(err, ErrResultsUnavailable):
		return httpError(http.StatusConflict,
			"only a table that has not ended can be opted in to the results feed: this one is a practice table, has ended, or is archived")
	default:
		return err
	}
}

// errResultsOffHTTP is the 503 the route answers on a server without a
// results channel, naming what turns it on (ADR 0174 §6).
var errResultsOffHTTP = httpError(http.StatusServiceUnavailable,
	"the results feed is off on this server: "+discord.ResultsChannelEnv+", "+
		discord.CommunityGuildEnv+" and "+discord.BotTokenEnv+" must all be set")
