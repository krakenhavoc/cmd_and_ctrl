// Package budget is #2980: a spend guard on the hosted model client.
//
// The `assisted` and `strong` bot tiers call a paid API when an
// Anthropic key is configured. A public launch must not be able to run
// up an unbounded bill, so every call to a hosted model goes through
// Budget.Wrap, which enforces three things in one place:
//
//   - a DAILY and a MONTHLY cap, counted in tokens from the usage the
//     API reports on each response, in UTC calendar days and months;
//   - an admin KILL SWITCH that turns model calls off without a
//     redeploy;
//   - persistence, so a restart (every deploy is one) does not hand a
//     capped server a fresh allowance.
//
// # What happens over the cap
//
// The same thing that happens with no key at all. The wrapped client
// returns model.ErrBudgetExceeded without sending anything;
// model.Policy treats that as FallbackBudgetCapped and plays Layer B's
// answer, which it computed before it asked. A seat already running
// keeps playing, on the heuristic. And the lobby's picker stops
// offering the model tiers, with Budget.Blocked as the reason: the
// same available:false-with-a-reason a keyless server gives (ADR 0033
// §6, amendment 2026-10-12).
//
// # Tokens, not dollars
//
// The cap counts tokens: input + output + cache reads + cache writes.
// The code has no price table, prices differ by model and change, and
// a dollar cap built on a stale table fails silently in whichever
// direction is worse. Tokens are what the response reports, so they
// need no table. The sum over-counts a little (a cache read is billed
// at a tenth of an input token and counted as one), which errs toward
// stopping early, the safe side for a cap.
//
// # What it deliberately does not do
//
// It does not reserve tokens before a call: with four seats, the
// calls in flight when the line is crossed can overrun it by a handful
// of replies a few hundred tokens long. It does not gate any tier
// behind payment, and never will: when a cap trips the model tiers
// fall back; no tier is ever sold.
//
// The cap applies to the HOSTED transport only. A self-hosted model
// has no bill, so main wires it with no caps (the kill switch still
// works).
package budget

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
)

// Default caps, in tokens. Conservative on purpose: a deployment that
// wants more says so in its env file; one that wants none sets 0.
//
// The figures are an order of magnitude, not a measurement. ADR 0033's
// per-game spend number is still pending a keyed run (#735), so what
// this pair buys is a ceiling: at a cheap hosted model's rates a
// million tokens is a few dollars at the very worst, and the month
// holds ten of those days.
const (
	DefaultDailyTokens   int64 = 1_000_000
	DefaultMonthlyTokens int64 = 10_000_000
)

// Limits are the caps. Zero means "no cap" for that window.
type Limits struct {
	DailyTokens   int64
	MonthlyTokens int64
}

// DefaultLimits is the shipped pair.
func DefaultLimits() Limits {
	return Limits{DailyTokens: DefaultDailyTokens, MonthlyTokens: DefaultMonthlyTokens}
}

// The env vars that set the caps. Unset takes the default; 0 turns
// that window's cap off; anything else must be a positive integer
// number of tokens, or the boot fails (a typo that silently left the
// default in place would be a cap the operator did not choose).
const (
	EnvDailyTokens   = "CMDCTRL_BOT_MODEL_DAILY_TOKENS"
	EnvMonthlyTokens = "CMDCTRL_BOT_MODEL_MONTHLY_TOKENS"
)

// LimitsFromEnv reads the two caps from the environment.
func LimitsFromEnv() (Limits, error) {
	return limitsFrom(os.Getenv)
}

func limitsFrom(getenv func(string) string) (Limits, error) {
	lim := DefaultLimits()
	for _, f := range []struct {
		name string
		dst  *int64
	}{{EnvDailyTokens, &lim.DailyTokens}, {EnvMonthlyTokens, &lim.MonthlyTokens}} {
		raw := strings.TrimSpace(getenv(f.name))
		if raw == "" {
			continue
		}
		n, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || n < 0 {
			return Limits{}, fmt.Errorf("%s=%q: want a whole number of tokens, or 0 for no cap", f.name, raw)
		}
		*f.dst = n
	}
	return lim, nil
}

// State is why model calls are or are not being made.
type State string

// The states, in the order they take priority.
const (
	StateOK         State = "ok"
	StateDisabled   State = "disabled"
	StateMonthlyCap State = "monthly_cap"
	StateDailyCap   State = "daily_cap"
)

// Snapshot is what a Store holds: the running totals for one UTC day
// and one UTC month, and the kill switch.
type Snapshot struct {
	DayTokens   int64
	MonthTokens int64
	Disabled    bool
}

// Store persists a Budget. Day and month are the period keys
// ("2026-10-12", "2026-10"); a store returns zero totals for a period
// it holds nothing for, which is how a roll-over reads after a
// restart.
type Store interface {
	Load(ctx context.Context, day, month string) (Snapshot, error)
	SaveTokens(ctx context.Context, day string, dayTokens int64, month string, monthTokens int64) error
	SaveDisabled(ctx context.Context, disabled bool) error
}

// Status is the readout: the admin endpoint's body and the lobby's
// note.
type Status struct {
	State State `json:"state"`
	// Enabled is false only when an admin switched model calls off.
	Enabled bool `json:"enabled"`
	// Note is the sentence a player sees. Empty when State is ok.
	Note string `json:"note,omitempty"`

	DailyTokens   int64 `json:"daily_tokens"`
	DailyLimit    int64 `json:"daily_limit"`
	MonthlyTokens int64 `json:"monthly_tokens"`
	MonthlyLimit  int64 `json:"monthly_limit"`
	// DayResetsAt and MonthResetsAt are when each counter next
	// starts from zero (UTC).
	DayResetsAt   time.Time `json:"day_resets_at"`
	MonthResetsAt time.Time `json:"month_resets_at"`
}

// Budget is the guard. Safe for concurrent use.
type Budget struct {
	mu    sync.Mutex
	lim   Limits
	store Store
	log   *slog.Logger
	now   func() time.Time

	day, month  string
	dayTokens   int64
	monthTokens int64
	disabled    bool
}

// Option adjusts a Budget at construction.
type Option func(*Budget)

// WithClock replaces the wall clock, for tests.
func WithClock(now func() time.Time) Option { return func(b *Budget) { b.now = now } }

// WithLogger sets the logger for persistence failures and cap events.
func WithLogger(l *slog.Logger) Option { return func(b *Budget) { b.log = l } }

// New builds a Budget and loads what the store holds. A nil store
// keeps everything in memory (a deployment with no data dir). A store
// that fails to load is logged and ignored: the guard then starts from
// zero, which is the lesser evil next to refusing to boot.
func New(lim Limits, store Store, opts ...Option) *Budget {
	b := &Budget{lim: lim, store: store, now: time.Now}
	for _, o := range opts {
		o(b)
	}
	if b.log == nil {
		b.log = slog.Default()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.day, b.month = periodKeys(b.now())
	if store != nil {
		snap, err := store.Load(context.Background(), b.day, b.month)
		if err != nil {
			b.log.Warn("bot model budget: could not load the persisted counters; starting from zero", "err", err)
		} else {
			b.dayTokens, b.monthTokens, b.disabled = snap.DayTokens, snap.MonthTokens, snap.Disabled
		}
	}
	return b
}

func periodKeys(t time.Time) (day, month string) {
	t = t.UTC()
	return t.Format("2006-01-02"), t.Format("2006-01")
}

// rollLocked starts a new day or month at the UTC boundary. The
// persisted rows of the old period are pruned by the next save.
func (b *Budget) rollLocked() {
	day, month := periodKeys(b.now())
	if day != b.day {
		b.day, b.dayTokens = day, 0
	}
	if month != b.month {
		b.month, b.monthTokens = month, 0
	}
}

func (b *Budget) stateLocked() State {
	switch {
	case b.disabled:
		return StateDisabled
	case b.lim.MonthlyTokens > 0 && b.monthTokens >= b.lim.MonthlyTokens:
		return StateMonthlyCap
	case b.lim.DailyTokens > 0 && b.dayTokens >= b.lim.DailyTokens:
		return StateDailyCap
	}
	return StateOK
}

// Check returns model.ErrBudgetExceeded when a call must not be made.
func (b *Budget) Check() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rollLocked()
	if b.stateLocked() != StateOK {
		return model.ErrBudgetExceeded
	}
	return nil
}

// Tokens is what one response counts against the caps.
func Tokens(u model.Usage) int64 {
	return int64(u.InputTokens) + int64(u.OutputTokens) + int64(u.CacheReadTokens) + int64(u.CacheWriteTokens)
}

// Record counts one response's usage and persists the totals.
func (b *Budget) Record(u model.Usage) {
	n := Tokens(u)
	if n <= 0 {
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rollLocked()
	before := b.stateLocked()
	b.dayTokens += n
	b.monthTokens += n
	if after := b.stateLocked(); after != StateOK && before == StateOK {
		b.log.Warn("bot model budget reached: the model tiers now play the heuristic until it resets",
			"state", string(after), "day_tokens", b.dayTokens, "month_tokens", b.monthTokens,
			"daily_limit", b.lim.DailyTokens, "monthly_limit", b.lim.MonthlyTokens)
	}
	if b.store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := b.store.SaveTokens(ctx, b.day, b.dayTokens, b.month, b.monthTokens); err != nil {
		b.log.Warn("bot model budget: could not persist the counters; a restart will forget this spend", "err", err)
	}
}

// SetEnabled is the kill switch. false turns model calls off until it
// is set back; the setting survives a restart.
func (b *Budget) SetEnabled(enabled bool) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.store != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := b.store.SaveDisabled(ctx, !enabled); err != nil {
			return fmt.Errorf("budget: persist the switch: %w", err)
		}
	}
	b.disabled = !enabled
	return nil
}

// Status is the current readout.
func (b *Budget) Status() Status {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.rollLocked()
	now := b.now().UTC()
	st := Status{
		State:         b.stateLocked(),
		Enabled:       !b.disabled,
		DailyTokens:   b.dayTokens,
		DailyLimit:    b.lim.DailyTokens,
		MonthlyTokens: b.monthTokens,
		MonthlyLimit:  b.lim.MonthlyTokens,
		DayResetsAt:   time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.UTC),
		MonthResetsAt: time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, time.UTC),
	}
	st.Note = noteFor(st)
	return st
}

// noteFor is the sentence for a player. It says what changed (the
// model tiers are off), what they get instead, and when it ends.
func noteFor(st Status) string {
	const instead = "Bots on the assisted and strong tiers are unavailable, and any already seated play like the heuristic tier. "
	switch st.State {
	case StateDisabled:
		return instead + "The server's operator has switched model calls off."
	case StateMonthlyCap:
		return instead + "This server's monthly model budget is used up; it resets " + st.MonthResetsAt.Format("2006-01-02 15:04") + " UTC."
	case StateDailyCap:
		return instead + "This server's daily model budget is used up; it resets " + st.DayResetsAt.Format("2006-01-02 15:04") + " UTC."
	}
	return ""
}

// Blocked is the note when model calls are not being made and empty
// when they are. The tier factory asks it for the picker's reason.
func (b *Budget) Blocked() string {
	st := b.Status()
	if st.State == StateOK {
		return ""
	}
	return st.Note
}

// Wrap guards a client. A nil client stays nil: no transport means no
// model tiers at all, which is a different (and older) state.
func (b *Budget) Wrap(c model.Client) model.Client {
	if c == nil {
		return nil
	}
	return &guarded{inner: c, b: b}
}

type guarded struct {
	inner model.Client
	b     *Budget
}

func (g *guarded) Name() string { return g.inner.Name() }

func (g *guarded) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := g.b.Check(); err != nil {
		return model.Response{}, err
	}
	resp, err := g.inner.Complete(ctx, req)
	g.b.Record(resp.Usage)
	return resp, err
}
