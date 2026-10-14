package lobby

// results_feed_test.go: ADR 0174 against an in-memory #game-results
// channel. Waits are on the channel's own signal and on WaitGameStats,
// never on a sleep.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/gamestats"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fakeResults is the #game-results channel: the posts the bot made.
type fakeResults struct {
	mu    sync.Mutex
	posts []discord.ResultsPost
	// errs are returned, one per call, before any post succeeds.
	errs []error
	// calls counts every attempt, failed ones included.
	calls int
	// posted is signalled once per successful post.
	posted chan struct{}
}

func newFakeResults(errs ...error) *fakeResults {
	return &fakeResults{errs: errs, posted: make(chan struct{}, 8)}
}

func (f *fakeResults) Post(_ context.Context, _ *http.Client, p discord.ResultsPost) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if len(f.errs) > 0 {
		err := f.errs[0]
		f.errs = f.errs[1:]
		if err != nil {
			return err
		}
	}
	f.posts = append(f.posts, p)
	f.posted <- struct{}{}
	return nil
}

func (f *fakeResults) snapshot() (posts []discord.ResultsPost, calls int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]discord.ResultsPost(nil), f.posts...), f.calls
}

func quietFeed(api resultsAPI) (*ResultsFeed, *bytes.Buffer) {
	var buf bytes.Buffer
	f := newResultsFeed(api, nil, slog.New(slog.NewTextHandler(&buf, nil)))
	f.wait = func(time.Duration) {}
	return f, &buf
}

// resultsLobby is a durable lobby with a fake channel attached and a
// two-seat table in play. minTurns 0 lets the test end it by a quick
// concession; the real rule is tested on buildResultsPost and by the
// concession test.
func resultsLobby(t *testing.T, optIn bool, minTurns int) (*Lobby, *fakeResults, GameMeta, uuid.UUID) {
	t.Helper()
	l, _ := newDurableLobby(t, t.TempDir())
	api := newFakeResults()
	feed, _ := quietFeed(api)
	feed.minTurns = minTurns
	l.SetResultsFeed(feed)
	meta, err := l.Create("Results night")
	if err != nil {
		t.Fatal(err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatal(err)
	}
	for id, name := range map[uuid.UUID]string{alice: "Alice's", bob: "Bob's"} {
		if _, err := l.SetDeck(meta.ID, id, name, botDeck(20)); err != nil {
			t.Fatal(err)
		}
	}
	if optIn {
		if m, err := l.SetResults(meta.ID, true); err != nil || !m.Results {
			t.Fatalf("SetResults: %v %v", m.Results, err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	return l, api, meta, alice
}

func concedeSeat(t *testing.T, l *Lobby, id, seat uuid.UUID) {
	t.Helper()
	room := l.RoomOf(id)
	if _, _, err := room.Apply(seat, func() error { return room.Game.Concede(seat) }); err != nil {
		t.Fatalf("conceding: %v", err)
	}
}

// awaitEnded waits until the lobby has seen the game end, which is when
// it starts the record and the post.
func awaitEnded(t *testing.T, l *Lobby, id uuid.UUID) {
	t.Helper()
	waitFor(t, func() bool { m, _ := l.Get(id); return m.State == string(game.StateEnded) })
}

func TestAnOptedInGameIsPostedWhenItEnds(t *testing.T) {
	l, api, meta, alice := resultsLobby(t, true, 0)
	concedeSeat(t, l, meta.ID, alice)
	select {
	case <-api.posted:
	case <-time.After(5 * time.Second):
		t.Fatal("no post within 5s")
	}
	l.WaitGameStats()
	posts, calls := api.snapshot()
	if len(posts) != 1 || calls != 1 {
		t.Fatalf("%d posts, %d calls; want exactly one", len(posts), calls)
	}
	p := posts[0]
	if p.Table != "Results night" || len(p.Seats) != 2 {
		t.Fatalf("post = %+v", p)
	}
	if p.Seats[0].Name != "Bob" || p.Seats[0].Place != 1 || p.Seats[0].Result != "win" {
		t.Errorf("first seat = %+v, want Bob the winner", p.Seats[0])
	}
	if p.Seats[1].Name != "Alice" || !p.Seats[1].Conceded || p.Seats[1].Place != 2 {
		t.Errorf("second seat = %+v, want Alice, conceded, 2nd", p.Seats[1])
	}
	if p.EndedAt.IsZero() || p.Duration < 0 {
		t.Errorf("times: ended %v, duration %v", p.EndedAt, p.Duration)
	}
	// Seeing the game end again (Get, List) posts nothing more.
	awaitEnded(t, l, meta.ID)
	l.WaitGameStats()
	if posts, _ := api.snapshot(); len(posts) != 1 {
		t.Errorf("%d posts after reading the ended game again, want 1", len(posts))
	}
}

func TestAGameIsNotPostedWithoutTheOptIn(t *testing.T) {
	l, api, meta, alice := resultsLobby(t, false, 0)
	concedeSeat(t, l, meta.ID, alice)
	awaitEnded(t, l, meta.ID)
	l.WaitGameStats()
	if posts, calls := api.snapshot(); len(posts) != 0 || calls != 0 {
		t.Errorf("%d posts, %d calls for a table that did not opt in", len(posts), calls)
	}
}

func TestAConcessionBeforeTurnThreeIsNotPosted(t *testing.T) {
	l, api, meta, alice := resultsLobby(t, true, resultsMinTurns)
	concedeSeat(t, l, meta.ID, alice)
	awaitEnded(t, l, meta.ID)
	l.WaitGameStats()
	if posts, calls := api.snapshot(); len(posts) != 0 || calls != 0 {
		t.Errorf("%d posts, %d calls for a game conceded in its first turns", len(posts), calls)
	}
}

func TestPracticeTablesAreNeverPosted(t *testing.T) {
	l, api, meta, _ := resultsLobby(t, true, 0)
	l.mu.Lock()
	entry := *l.games[meta.ID]
	l.mu.Unlock()
	entry.practice = true
	l.mu.Lock()
	job := l.resultsJobLocked(&entry)
	l.recordStatsLocked(&entry)
	l.mu.Unlock()
	l.WaitGameStats()
	if job != nil {
		t.Error("a practice table got a results job")
	}
	if _, calls := api.snapshot(); calls != 0 {
		t.Errorf("a practice table made %d calls", calls)
	}
}

func TestSetResults(t *testing.T) {
	l := newTestLobby(t)
	meta, err := l.Create("T")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := l.SetResults(meta.ID, true); !errors.Is(err, ErrResultsOff) {
		t.Errorf("with the feed off: %v", err)
	}
	if l.ResultsEnabled() {
		t.Error("ResultsEnabled true with no feed")
	}
	feed, _ := quietFeed(newFakeResults())
	l.SetResultsFeed(feed)
	if !l.ResultsEnabled() {
		t.Error("ResultsEnabled false with a feed attached")
	}
	if _, err := l.SetResults(uuid.New(), true); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("unknown table: %v", err)
	}
	if m, err := l.SetResults(meta.ID, true); err != nil || !m.Results {
		t.Fatalf("on: %v %v", m.Results, err)
	}
	if m, err := l.SetResults(meta.ID, true); err != nil || !m.Results {
		t.Fatalf("on again: %v %v", m.Results, err)
	}
	if m, err := l.SetResults(meta.ID, false); err != nil || m.Results {
		t.Fatalf("off: %v %v", m.Results, err)
	}
}

func TestTheOptInSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	feed, _ := quietFeed(newFakeResults())
	l.SetResultsFeed(feed)
	meta, err := l.Create("Durable")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"Alice", "Bob"} {
		if _, _, err := l.Join(meta.ID, meta.InviteToken, name); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.SetResults(meta.ID, true); err != nil {
		t.Fatal(err)
	}
	l2, _ := newDurableLobby(t, dir)
	l2.RestoreFromDisk(quietLogger())
	got, err := l2.Get(meta.ID)
	if err != nil || !got.Results {
		t.Fatalf("after a restart: results %v, err %v", got.Results, err)
	}
}

func TestAnEndedGameCannotOptIn(t *testing.T) {
	l, api, ended, alice := resultsLobby(t, false, 0)
	concedeSeat(t, l, ended.ID, alice)
	awaitEnded(t, l, ended.ID)
	l.WaitGameStats()
	if _, err := l.SetResults(ended.ID, true); !errors.Is(err, ErrResultsUnavailable) {
		t.Errorf("opting an ended game in: %v", err)
	}
	if m, err := l.SetResults(ended.ID, false); err != nil || m.Results {
		t.Errorf("opting an ended game out: %v %v", m.Results, err)
	}
	if posts, _ := api.snapshot(); len(posts) != 0 {
		t.Error("an ended game opted in too late was posted")
	}
}

// --- the rules, on a record built by hand -----------------------------------

func placeInt(n int) *int { return &n }

func handRecord(turns int, conceded bool) (protocol.GameView, gamestats.Game) {
	view := protocol.GameView{Seats: []protocol.PlayerView{
		{ID: "a", Name: "Alice", Seat: 0},
		{ID: "g", Name: "Gus", Seat: 1},
		{ID: "c", Name: "Claude", Seat: 2, IsAgent: true},
		{ID: "b", Name: "Bot", Seat: 3, IsBot: true},
	}}
	rec := gamestats.Game{Turns: turns, Seats: []gamestats.Seat{
		{Seat: 0, Result: gamestats.ResultLoss, Place: placeInt(3), Conceded: conceded, EliminatedTurn: placeInt(turns), Commanders: []string{"Atraxa"}},
		{Seat: 1, Result: gamestats.ResultWin, Place: placeInt(1), Commanders: []string{"Tymna", "Thrasios"}},
		{Seat: 2, Result: gamestats.ResultLoss, Place: placeInt(2), EliminatedTurn: placeInt(turns)},
		{Seat: 3, Result: gamestats.ResultLoss, Place: placeInt(4), EliminatedTurn: placeInt(turns)},
	}}
	return view, rec
}

func TestBuildResultsPostRules(t *testing.T) {
	feed, _ := quietFeed(newFakeResults())
	job := func(outcome string) *resultsJob {
		return &resultsJob{feed: feed, table: "Friday", outcome: outcome,
			from: time.Unix(0, 0), to: time.Unix(0, 0).Add(90 * time.Minute)}
	}
	view, rec := handRecord(9, false)
	p, ok := buildResultsPost(job(game.OutcomeWin), view, rec)
	if !ok {
		t.Fatal("a played-out game was not posted")
	}
	wantOrder := []string{"Gus", "Claude", "Alice", "Bot"}
	for i, s := range p.Seats {
		if s.Name != wantOrder[i] {
			t.Errorf("seat %d = %s, want %s (placing order)", i, s.Name, wantOrder[i])
		}
	}
	if p.Seats[1].Kind != "AI agent" || p.Seats[3].Kind != "bot" || p.Seats[0].Kind != "" {
		t.Errorf("kinds = %q %q %q", p.Seats[0].Kind, p.Seats[1].Kind, p.Seats[3].Kind)
	}
	if p.Turns != 9 || p.Duration != 90*time.Minute || len(p.Seats[0].Commanders) != 2 {
		t.Errorf("post = %+v", p)
	}

	if _, ok := buildResultsPost(job(""), view, rec); ok {
		t.Error("a game with no outcome (closed, not played) was posted")
	}
	short, shortRec := handRecord(2, true)
	if _, ok := buildResultsPost(job(game.OutcomeWin), short, shortRec); ok {
		t.Error("a concession on turn 2 was posted")
	}
	late, lateRec := handRecord(3, true)
	if _, ok := buildResultsPost(job(game.OutcomeWin), late, lateRec); !ok {
		t.Error("a concession on turn 3 was not posted")
	}
	// A short game nobody conceded is still a result.
	quick, quickRec := handRecord(2, false)
	if _, ok := buildResultsPost(job(game.OutcomeWin), quick, quickRec); !ok {
		t.Error("a short game with no concession was not posted")
	}
}

func TestPostRetriesOnlyTransientFailures(t *testing.T) {
	id := uuid.New()
	p := discord.ResultsPost{Table: "T"}

	api := newFakeResults(discord.ErrVoiceUnavailable, &discord.RateLimitError{RetryAfter: time.Hour})
	feed, _ := quietFeed(api)
	var waits []time.Duration
	feed.wait = func(d time.Duration) { waits = append(waits, d) }
	feed.post(id, p)
	if posts, calls := api.snapshot(); len(posts) != 1 || calls != 3 {
		t.Errorf("%d posts in %d calls; want a post on the third", len(posts), calls)
	}
	if len(waits) != 2 || waits[0] != resultsBackoff || waits[1] != resultsMaxRetryWait {
		t.Errorf("waits = %v; want the backoff, then the 429's wait capped", waits)
	}

	api = newFakeResults(discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, discord.ErrVoiceUnavailable, nil)
	feed, buf := quietFeed(api)
	feed.post(id, p)
	if posts, calls := api.snapshot(); len(posts) != 0 || calls != resultsMaxAttempts {
		t.Errorf("%d posts in %d calls; want to give up after %d", len(posts), calls, resultsMaxAttempts)
	}
	if buf.Len() == 0 {
		t.Error("a dropped post was not logged")
	}

	api = newFakeResults(discord.ErrResultsForbidden)
	feed, _ = quietFeed(api)
	feed.post(id, p)
	if _, calls := api.snapshot(); calls != 1 {
		t.Errorf("a 403 was tried %d times", calls)
	}
}

func TestResultsRidesTheMetaJSON(t *testing.T) {
	raw, err := json.Marshal(GameMeta{Results: true})
	if err != nil || !bytes.Contains(raw, []byte(`"results":true`)) {
		t.Errorf("meta JSON = %s, %v", raw, err)
	}
	raw, _ = json.Marshal(GameMeta{})
	if bytes.Contains(raw, []byte(`"results"`)) {
		t.Errorf("an opted-out meta carries results: %s", raw)
	}
}
