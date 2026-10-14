package notify

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// The timings of ADR 0162 §2.
const (
	// TurnWait is how long the engine must have waited on a seat, with
	// the seat away the whole time, before "your turn" is sent. Owner
	// answer 2.
	TurnWait = 2 * time.Minute
	// Window is the rate limit: at most one DM per player per table per
	// Window, across every kind.
	Window = 10 * time.Minute
	// EventGrace is how long a "filled" or "started" waits before it is
	// looked at. A player who has just joined, or just pressed Start,
	// may not have opened the table's socket yet; by the time the grace
	// is up they have, and are not away.
	EventGrace = 30 * time.Second
	// TickEvery is how often Run looks at the tables. The clocks are
	// only as fine as this, and always err late.
	TickEvery = 15 * time.Second
	// sendTimeout bounds one notification: the recipient lookup, the
	// two Discord calls, a 429's one wait and the second try.
	sendTimeout = 30 * time.Second
	// maxEvents bounds the queue of unprocessed table events. Past it
	// an event is dropped: a missed "table ready" costs less than an
	// unbounded queue.
	maxEvents = 1024
)

// Table states, as Tables reports them.
const (
	StateLobby  = "lobby"
	StateActive = "active"
)

// Seat is one seat at a table, as the Notifier needs it.
type Seat struct {
	PlayerID uuid.UUID
	// UserID is the signed-in person holding the seat, uuid.Nil for a
	// guest, a bot, an agent and a seat still waiting for its person.
	UserID uuid.UUID
	Host   bool
	Bot    bool
	Agent  bool
}

// person reports whether the seat may ever be messaged: a signed-in
// person, never a bot, an agent seat (ADR 0122) or a guest.
func (s Seat) person() bool { return s.UserID != uuid.Nil && !s.Bot && !s.Agent }

// Table is one table, as the Notifier needs it.
type Table struct {
	ID   uuid.UUID
	Name string
	// State is StateLobby, StateActive, or anything else for a table
	// that is neither (ended), which nothing is sent about.
	State string
	// Turn is the turn number the board shows, 0 before turn 1.
	Turn  int
	Seats []Seat
	// Full is whether every seat is taken.
	Full bool
}

// Tables is the lobby's half: which tables there are, and who the
// engine is waiting on. Practice tables are never reported.
type Tables interface {
	// Live lists the tables in play, not archived and not practice.
	Live() []Table
	// Table looks one table up; false for an unknown or practice table.
	Table(id uuid.UUID) (Table, bool)
	// Waiting reports whether the engine is waiting on seat for a
	// decision, and when seat last committed an action of its own (zero
	// for never).
	Waiting(game, seat uuid.UUID) (waiting bool, lastActed time.Time)
}

// Presence is the hub's half: ws.Hub.SeatAway.
type Presence interface {
	// SeatAway reports whether the seat is away, and since when (zero
	// when the hub has not seen it leave).
	SeatAway(game, seat uuid.UUID) (away bool, since time.Time)
}

// Recipients turns a user into the Discord snowflake to message.
// users.Store satisfies it.
type Recipients interface {
	DiscordSubject(ctx context.Context, user uuid.UUID) (string, error)
}

// Sender sends one DM. BotSender is the production one.
type Sender interface {
	SendDM(ctx context.Context, recipient, content string) error
}

// BotSender sends through the server's bot token.
type BotSender struct {
	Bot    discord.Bot
	Client *http.Client
}

// SendDM implements Sender.
func (b BotSender) SendDM(ctx context.Context, recipient, content string) error {
	return b.Bot.SendDM(ctx, b.Client, recipient, content)
}

// Config is what a Notifier is built from. Every interface is required.
type Config struct {
	Tables     Tables
	Presence   Presence
	Store      Store
	Recipients Recipients
	Sender     Sender
	// BaseURL is the public origin the link is built against
	// (CMDCTRL_PUBLIC_BASE_URL).
	BaseURL string
	Log     *slog.Logger
	// Now and Sleep are the clock; nil is the real one. A test sets
	// both.
	Now   func() time.Time
	Sleep func(ctx context.Context, d time.Duration) error
}

// seatKey names one seat at one table.
type seatKey struct{ game, seat uuid.UUID }

// windowKey names one player at one table, for the rate window.
type windowKey struct{ user, game uuid.UUID }

// clock is one seat's "your turn" clock.
type clock struct {
	// since is when this Notifier first saw the engine waiting on the
	// seat, in this unbroken stretch of waiting.
	since time.Time
	// firedAt is when "your turn" last went out (or was held back by
	// the window) for this stretch, zero for not yet. Acting or coming
	// back starts a new stretch, which may fire again.
	firedAt time.Time
}

type eventKind int

const (
	eventFilled eventKind = iota
	eventStarted
)

// event is a table filling or starting, queued for EventGrace.
type event struct {
	kind eventKind
	game uuid.UUID
	// by is the seat whose join filled the table: it is the one seat
	// that is certainly not away, and is never told.
	by  uuid.UUID
	due time.Time
}

// Notifier decides who to message and sends it. Safe for concurrent
// use: the lobby queues events from its request goroutines while Run
// ticks on its own.
type Notifier struct {
	cfg Config
	log *slog.Logger

	mu     sync.Mutex
	events []event
	clocks map[seatKey]*clock
	// awaySeen is when this Notifier first saw a seat away that the
	// hub has no departure time for (it has not been connected since
	// the server started).
	awaySeen map[seatKey]time.Time
	// lastSent is the rate window: when a DM last went to a player
	// about a table. In memory only; a restart forgets it (ADR 0162
	// §5), which at worst sends one extra DM.
	lastSent map[windowKey]time.Time
}

// New builds a Notifier. It does nothing until Run, or a test's Tick.
func New(cfg Config) *Notifier {
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.Sleep == nil {
		cfg.Sleep = sleep
	}
	return &Notifier{
		cfg:      cfg,
		log:      log,
		clocks:   make(map[seatKey]*clock),
		awaySeen: make(map[seatKey]time.Time),
		lastSent: make(map[windowKey]time.Time),
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Run ticks every TickEvery until ctx ends.
func (n *Notifier) Run(ctx context.Context) {
	t := time.NewTicker(TickEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			n.Tick(ctx)
		}
	}
}

// TableFilled queues "the table filled" for table game. by is the seat
// whose join filled it (uuid.Nil for a bot), which is never told.
func (n *Notifier) TableFilled(game, by uuid.UUID) {
	n.queue(event{kind: eventFilled, game: game, by: by})
}

// TableStarted queues "the table started" for table game.
func (n *Notifier) TableStarted(game uuid.UUID) {
	n.queue(event{kind: eventStarted, game: game})
}

func (n *Notifier) queue(ev event) {
	ev.due = n.cfg.Now().Add(EventGrace)
	n.mu.Lock()
	defer n.mu.Unlock()
	if len(n.events) >= maxEvents {
		return
	}
	n.events = append(n.events, ev)
}

// Tick does one round: the table events that are due, then every
// opted-in seat's turn clock. Exported for tests, which drive it with a
// fake clock instead of Run.
func (n *Notifier) Tick(ctx context.Context) {
	now := n.cfg.Now()
	active, err := n.cfg.Store.Active(ctx)
	if err != nil {
		n.log.Warn("notify: reading the notification switches failed", "err", err)
		return
	}
	n.pruneWindow(now)
	n.tickEvents(ctx, now, active)
	n.tickTurns(ctx, now, active)
}

// pruneWindow forgets rate-window entries that no longer limit
// anything.
func (n *Notifier) pruneWindow(now time.Time) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for k, at := range n.lastSent {
		if now.Sub(at) >= Window {
			delete(n.lastSent, k)
		}
	}
}

// claim reports whether user may be sent a DM about game now, and if
// so takes the window. The window is taken before the send, whatever
// the send's outcome: a notification is never retried late (§4).
func (n *Notifier) claim(user, game uuid.UUID, now time.Time) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	k := windowKey{user, game}
	if at, ok := n.lastSent[k]; ok && now.Sub(at) < Window {
		return false
	}
	n.lastSent[k] = now
	return true
}

// tickEvents handles every queued event whose grace is up.
func (n *Notifier) tickEvents(ctx context.Context, now time.Time, active map[uuid.UUID]Prefs) {
	n.mu.Lock()
	var due []event
	kept := n.events[:0]
	for _, ev := range n.events {
		if now.Before(ev.due) {
			kept = append(kept, ev)
		} else {
			due = append(due, ev)
		}
	}
	n.events = kept
	n.mu.Unlock()

	for _, ev := range due {
		t, ok := n.cfg.Tables.Table(ev.game)
		if !ok {
			continue
		}
		switch ev.kind {
		case eventFilled:
			// Only while it is still full and waiting to start: once it
			// has started, the start's own event says so.
			if t.State != StateLobby || !t.Full {
				continue
			}
		case eventStarted:
			if t.State != StateActive {
				continue
			}
		}
		for _, s := range t.Seats {
			if !s.person() || s.PlayerID == ev.by {
				continue
			}
			p, ok := active[s.UserID]
			if !ok {
				continue
			}
			kind, ok := eventKindFor(ev.kind, s, p)
			if !ok {
				continue
			}
			if away, _ := n.cfg.Presence.SeatAway(t.ID, s.PlayerID); !away {
				continue
			}
			if !n.claim(s.UserID, t.ID, now) {
				continue
			}
			var content string
			if ev.kind == eventFilled {
				content = filledMessage(t.Name, n.link(t.ID))
			} else {
				content = startedMessage(t.Name, n.link(t.ID))
			}
			n.send(ctx, now, active, s.UserID, t.ID, kind, content)
		}
	}
}

// eventKindFor is which switch an event reaches a seat through. A host
// is told a fill as "table filled" when that switch is on, and as
// "table ready" otherwise; everyone else only ever as "table ready".
func eventKindFor(k eventKind, s Seat, p Prefs) (Kind, bool) {
	if k == eventFilled && s.Host && p.Wants(KindTableFilled) {
		return KindTableFilled, true
	}
	if p.Wants(KindTableReady) {
		return KindTableReady, true
	}
	return "", false
}

// tickTurns runs every opted-in seat's turn clock.
func (n *Notifier) tickTurns(ctx context.Context, now time.Time, active map[uuid.UUID]Prefs) {
	type due struct {
		t    Table
		user uuid.UUID
	}
	var send []due
	seen := make(map[seatKey]bool)
	for _, t := range n.cfg.Tables.Live() {
		if t.State != StateActive {
			continue
		}
		for _, s := range t.Seats {
			if !s.person() {
				continue
			}
			if p, ok := active[s.UserID]; !ok || !p.Wants(KindTurn) {
				continue
			}
			k := seatKey{t.ID, s.PlayerID}
			seen[k] = true
			waiting, acted := n.cfg.Tables.Waiting(t.ID, s.PlayerID)
			away, left := n.cfg.Presence.SeatAway(t.ID, s.PlayerID)
			if n.turnDue(k, now, waiting, acted, away, left) && n.claim(s.UserID, t.ID, now) {
				send = append(send, due{t, s.UserID})
			}
		}
	}
	n.mu.Lock()
	for k := range n.clocks {
		if !seen[k] {
			delete(n.clocks, k)
		}
	}
	for k := range n.awaySeen {
		if !seen[k] {
			delete(n.awaySeen, k)
		}
	}
	n.mu.Unlock()

	for _, d := range send {
		n.send(ctx, now, active, d.user, d.t.ID, KindTurn, turnMessage(d.t.Name, d.t.Turn, n.link(d.t.ID)))
	}
}

// turnDue advances one seat's clock and reports whether "your turn" is
// due now. The stretch the two minutes are counted over starts at the
// latest of: when the waiting began, when the seat last acted, and when
// the seat went away. So acting, or coming back, starts it again.
func (n *Notifier) turnDue(k seatKey, now time.Time, waiting bool, acted time.Time, away bool, left time.Time) bool {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !waiting {
		delete(n.clocks, k)
		return false
	}
	c := n.clocks[k]
	if c == nil {
		c = &clock{since: now}
		n.clocks[k] = c
	}
	if !away {
		delete(n.awaySeen, k)
		return false
	}
	if left.IsZero() {
		seen, ok := n.awaySeen[k]
		if !ok {
			seen = now
			n.awaySeen[k] = now
		}
		left = seen
	}
	start := latest(c.since, acted, left)
	if now.Sub(start) < TurnWait {
		return false
	}
	if !c.firedAt.IsZero() && !start.After(c.firedAt) {
		return false
	}
	c.firedAt = now
	return true
}

func latest(ts ...time.Time) time.Time {
	var out time.Time
	for _, t := range ts {
		if t.After(out) {
			out = t
		}
	}
	return out
}

// send delivers one DM and counts it. A 429 waits once (Discord's
// retry_after, capped) and tries again; anything else is logged and
// dropped. 50007 records the refusal, which stops every later DM to the
// person until they turn a switch on again.
func (n *Notifier) send(ctx context.Context, now time.Time, active map[uuid.UUID]Prefs, user, game uuid.UUID, kind Kind, content string) {
	if _, ok := active[user]; !ok {
		// Refused earlier in this same tick.
		return
	}
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	recipient, err := n.cfg.Recipients.DiscordSubject(ctx, user)
	if err != nil {
		n.log.Warn("notify: no Discord account to message", "kind", string(kind), "game", game.String(), "err", err)
		metrics.DMNotification(string(kind), metrics.DMError)
		return
	}
	err = n.cfg.Sender.SendDM(ctx, recipient, content)
	if errors.Is(err, discord.ErrDMRateLimited) {
		wait := discord.RetryAfter(err)
		if wait <= 0 {
			wait = time.Second
		}
		if n.cfg.Sleep(ctx, wait) == nil {
			err = n.cfg.Sender.SendDM(ctx, recipient, content)
		}
	}
	switch {
	case err == nil:
		metrics.DMNotification(string(kind), metrics.DMSent)
	case errors.Is(err, discord.ErrDMRefused):
		delete(active, user)
		if markErr := n.cfg.Store.MarkFailed(ctx, user, now); markErr != nil {
			n.log.Warn("notify: recording a refused DM failed", "err", markErr)
		}
		n.log.Info("notify: Discord refused a DM; no more until the person turns a switch on again",
			"kind", string(kind), "user_id", user.String())
		metrics.DMNotification(string(kind), metrics.DMRefused)
	default:
		n.log.Warn("notify: sending a DM failed; dropped", "kind", string(kind), "game", game.String(), "err", err)
		metrics.DMNotification(string(kind), metrics.DMError)
	}
}

// link is the table's address: the client's game route, which rejoins a
// signed-in player's seat through POST /me/games/{id}/session (ADR
// 0112).
func (n *Notifier) link(game uuid.UUID) string {
	return gameURL(n.cfg.BaseURL, game)
}
