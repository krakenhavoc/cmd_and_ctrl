// Package botarena plays headless bot-vs-bot games and reports what
// happened — ADR 0052 decision 4, the measurement half of the bot
// eval harness.
//
// # Why this is not a test
//
// aiseat's whole-game tests already seat four bots and play to a
// winner. What they cannot do is answer "is `assisted` stronger than
// `heuristic`", because that question needs tens of games, minutes of
// wall clock, a model endpoint, a Scryfall dump, and an answer that
// is a number with a confidence interval rather than a pass or a
// fail. This package is that: a library the `boteval arena`
// subcommand drives, whose output is a report somebody reads.
//
// # Why it lives outside aiseat/
//
// An arena has to hold the authoritative *game.Game and a *ws.Room.
// Every package under aiseat/ is forbidden from importing
// internal/game (ADR 0033 §3, enforced by
// heuristic/imports_test.go over Imports, TestImports AND
// XTestImports), because a POLICY holding authoritative state could
// read an opponent's hand. The ban is on policies; this is a harness,
// and it hands each policy nothing but the filtered aiseat.Input the
// runner would. So it sits here, above the ban, next to the engine.
//
// # What the numbers mean
//
// A win rate from N games is a coin-flip estimate and is reported as
// one: every rate carries a Wilson 95% interval and the null rate the
// table would produce by chance (1/seats). Four-seat tables rotate
// seat order between games, because turn order in Commander is worth
// real percentage points and a fixed order would measure the chair
// rather than the policy.
package botarena

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/deckprofile"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/rules"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/tiers"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// SeatSpec is one contestant: a tier, the deck it plays, and the name
// its results are tallied under.
type SeatSpec struct {
	// Tier is the bot difficulty (random, heuristic, assisted, strong).
	Tier tiers.Tier `json:"tier"`
	// Deck is a curated deck id. Empty deals BattleDeck, which needs
	// no Scryfall dump and is what the ungated tests use.
	Deck string `json:"deck,omitempty"`
	// Name is the key this seat's results are tallied under. Empty
	// takes the tier's own name, which is what a plain
	// "assisted vs heuristic" run wants; set it to compare two
	// configurations of the SAME tier in one run.
	Name string `json:"name,omitempty"`
}

// Label is the tally key: Name, or the tier.
func (s SeatSpec) Label() string {
	if strings.TrimSpace(s.Name) != "" {
		return s.Name
	}
	return string(s.Tier)
}

// Config is one arena run.
type Config struct {
	// Seats are the contestants, in their game-0 order. At least two.
	Seats []SeatSpec
	// Games is how many to play. At least one.
	Games int
	// Seed is the first game's seed; game i uses Seed+i, so a run is
	// exactly reproducible and two runs with different policies can
	// be played on the same deals.
	Seed uint64
	// Rotate moves each spec one chair along per game: spec k sits at
	// position (k+i)%n in game i. On with more than two seats it is
	// the difference between measuring a policy and measuring a
	// chair. Off replays the same seating every game, which is what
	// you want when isolating a single change.
	Rotate bool
	// TurnBudget stops a game that will not end. Default 60.
	TurnBudget int
	// Wall is the per-game deadline. Default 30m — generous, because
	// four model seats on one GPU serialise.
	Wall time.Duration
	// Stall is how long the room's sequence number may stand still
	// before the game is declared stalled. Default 3×MaxThink+15s.
	Stall time.Duration
	// Index is the Scryfall index curated decks and model deck
	// profiles are built from. Nil is fine for BattleDeck runs.
	Index *cards.Index
	// Client is the model transport. Nil refuses a model tier rather
	// than quietly seating Layer A + B under its name — see validate.
	Client model.Client
	// Models overrides the two model ids the funnel asks for.
	Models tiers.Models
	// MaxThink is the model tiers' per-window deadline.
	MaxThink time.Duration
	// ModelConfig overrides the funnel's tuning. Nil takes the
	// tier's default. This is the knob every Part 2 lever is A/B'd on.
	ModelConfig *model.Config
	// Runner is the runner pacing. The ZERO VALUE is the arena
	// default and is deliberately not aiseat.DefaultConfig: MinThink
	// 0 (nobody is watching, so there is no reason to hold a decision
	// back) and Narrate off (there is no chat to narrate to). MaxThink
	// is filled per seat from the tier when left at zero.
	//
	// BlockGrace is the ONE field where zero does not mean the arena
	// default: it is filled with PRODUCTION's 4s. It is how long an
	// attacking runner holds its pass during declare-blockers while
	// some defender still has a legal block, and aiseat's own
	// withDefaults does not fill it. With MinThink at 0 and no grace
	// the attacker re-steps on its own commit and races the
	// defenders, so every combat resolves with systematically fewer
	// blocks than the same policies would declare at a real table —
	// and combat is where policies differ, so the bias lands squarely
	// on the one number this harness exists to produce. A NEGATIVE
	// duration turns the hold off (aiseat treats <= 0 as off), which
	// is the only way to buy back the wall clock it costs;
	// `boteval arena --block-grace 0` spells exactly that.
	Runner aiseat.Config
	// DecisionLog, when set, gets one log file per game.
	DecisionLog *decisionlog.Logger
	// ReplayDir, when set, turns on the room's JSONL replay. Off by
	// default: a four-seat game's replay is ~320 MiB.
	ReplayDir string
	// Log is where the harness and the runners log. Nil discards.
	Log *slog.Logger
	// Note is free-form text recorded in the report — the model's
	// quantisation, what was being tested, whatever the next reader
	// will wish somebody had written down.
	Note string
}

const (
	defaultTurnBudget = 60
	defaultWall       = 30 * time.Minute
	defaultStallBase  = 15 * time.Second
)

func (c Config) withDefaults() Config {
	if c.TurnBudget <= 0 {
		c.TurnBudget = defaultTurnBudget
	}
	if c.Wall <= 0 {
		c.Wall = defaultWall
	}
	if c.Stall <= 0 {
		c.Stall = 3*c.MaxThink + defaultStallBase
	}
	if c.Runner.BlockGrace == 0 {
		// See Config.Runner. Zero is production's 4s, negative is
		// off; aiseat.Config.withDefaults fills neither.
		c.Runner.BlockGrace = aiseat.DefaultConfig().BlockGrace
	}
	if c.Log == nil {
		c.Log = slog.New(slog.NewTextHandler(discard{}, &slog.HandlerOptions{Level: slog.LevelError}))
	}
	return c
}

// discard is an io.Writer that drops everything, so that the nil-Log
// default costs one method call rather than a nil check at every use.
type discard struct{}

func (discard) Write(p []byte) (int, error) { return len(p), nil }

// Validate reports the configuration errors worth refusing before a
// single game is played.
func (c Config) Validate() error {
	if len(c.Seats) < 2 {
		return fmt.Errorf("botarena: need at least 2 seats, got %d", len(c.Seats))
	}
	if len(c.Seats) > game.MaxPlayers {
		return fmt.Errorf("botarena: %d seats, the engine seats at most %d", len(c.Seats), game.MaxPlayers)
	}
	if c.Games < 1 {
		return fmt.Errorf("botarena: need at least 1 game, got %d", c.Games)
	}
	for i, s := range c.Seats {
		if _, err := tiers.Parse(string(s.Tier)); err != nil {
			return fmt.Errorf("botarena: seat %d: %w", i, err)
		}
		// A model tier with no client is the one misconfiguration
		// that would silently corrupt the measurement rather than
		// break it: tiers.New happily builds an `assisted` seat with
		// a nil Client, it plays Layer A + B, and the report says
		// "assisted won 51%" about a heuristic. The lobby refuses the
		// same thing (docs/bot.md, "An unavailable tier is refused,
		// not downgraded"); so does this.
		if s.Tier.NeedsModel() && c.Client == nil {
			return fmt.Errorf("botarena: seat %d is %s, which needs a model endpoint: set --endpoint or $CMDCTRL_OPENAI_ENDPOINT (an %s seat with no client silently plays the heuristic)", i, s.Tier, s.Tier)
		}
		if s.Deck != "" && c.Index == nil {
			return fmt.Errorf("botarena: seat %d plays %q, which needs a Scryfall dump: set --dump or $CMDCTRL_SCRYFALL_DUMP", i, s.Deck)
		}
	}
	return nil
}

// Order is the seating for game i: Order(n, i, rotate)[p] is the
// index into Config.Seats of the spec sitting at position p.
//
// The rotation is "spec k sits at position (k+i) mod n", which is the
// cyclic shift that gives every spec every chair exactly Games/n
// times over a run whose length is a multiple of n. Anything fancier
// (a random permutation per game) would need many more games to
// balance and could not be reasoned about from the seed alone.
func Order(n, game int, rotate bool) []int {
	out := make([]int, n)
	for k := 0; k < n; k++ {
		pos := k
		if rotate {
			pos = ((k+game)%n + n) % n
		}
		out[pos] = k
	}
	return out
}

// ChairCounts is the seating a run will actually produce:
// ChairCounts(n, games, rotate)[k][p] is how many of the games spec k
// spends in chair p.
//
// It is recorded with every run because rotation only balances over a
// run whose length is a multiple of the table size, and the residual
// is a turn-order bias of the same order as the effect being
// measured. A finished run has to be auditable on that point rather
// than trusted.
func ChairCounts(n, games int, rotate bool) [][]int {
	if n <= 0 || games <= 0 {
		return nil
	}
	out := make([][]int, n)
	for k := range out {
		out[k] = make([]int, n)
	}
	for i := 0; i < games; i++ {
		for pos, k := range Order(n, i, rotate) {
			out[k][pos]++
		}
	}
	return out
}

// ChairBalanceWarning names the turn-order bias a rotated run cannot
// cancel, or "" when there is none.
//
// Rotation moves each contestant one chair along per game, so it
// balances exactly when the run length is a whole number of
// rotations. The default `--games 10` on a four-seat table is not:
// two contestants get three first-chair games and two get two. Turn
// order in Commander is worth real percentage points, which is the
// same order as the difference an arena run is trying to resolve.
//
// It is a WARNING and not a correction on purpose. Quietly rounding
// the operator's --games up to 12 would spend three unasked-for hours
// of GPU time; quietly rounding it down to 8 would throw two games of
// evidence away. Saying so, and putting the chair histogram in the
// report, leaves the choice where it belongs.
func ChairBalanceWarning(n, games int, rotate bool) string {
	if !rotate || n <= 1 || games <= 0 || games%n == 0 {
		return ""
	}
	extra := games % n
	low := games / n
	lower, upper := games-extra, games-extra+n
	if lower == 0 {
		lower, upper = upper, upper+n
	}
	return fmt.Sprintf("rotation cannot balance turn order: %d games over %d seats is not a whole number of rotations, so %d of the %d contestants sit in each chair %d times and the other %d sit there %d times. Turn order in Commander is worth real percentage points — the same order as the difference being measured. Play %d or %d games for a balanced run; the chair histogram is in the report either way.",
		games, n, extra, n, low+1, n-extra, low, lower, upper)
}

// SeatResult is one spec's game: where it sat, whether it survived,
// and everything its runner and its funnel counted.
type SeatResult struct {
	Spec     SeatSpec `json:"spec"`
	Position int      `json:"position"`
	// Policy is the name the POLICY reported, which is not
	// necessarily the tier that was asked for. A seat whose tier and
	// policy disagree is the one bug this layer cannot otherwise
	// show you, so it is recorded rather than assumed.
	Policy     string           `json:"policy"`
	Life       int              `json:"life"`
	Eliminated bool             `json:"eliminated"`
	Runner     aiseat.Stats     `json:"runner"`
	Funnel     *model.Stats     `json:"funnel,omitempty"`
	Meter      rules.MeterStats `json:"meter"`
	// Decision and ModelCall are this seat's own latency
	// distributions for this game, measured from the runner's
	// decision events rather than from the runner's internal ring —
	// the ring holds the most recent 1024 windows and a long game
	// rolls off the front.
	Decision  aiseat.Percentiles `json:"decision_latency"`
	ModelCall aiseat.Percentiles `json:"model_call_latency"`
	// PromptBytesP50 is the median rendered prompt size, zero when
	// this seat assembled no prompts.
	PromptBytesP50 int `json:"prompt_bytes_p50,omitempty"`

	// raw carries the samples the two distributions above were
	// computed from, so that Run can take a run-wide percentile
	// rather than a percentile of percentiles. Never serialised: a
	// games.jsonl with every latency in it would be enormous and
	// nobody reads it.
	raw *samples
}

// GameResult is one game.
type GameResult struct {
	Seed   uint64       `json:"seed"`
	GameID uuid.UUID    `json:"game_id"`
	Order  []int        `json:"order"`
	Seats  []SeatResult `json:"seats"`
	Turns  int          `json:"turns"`
	State  game.State   `json:"state"`
	// Winner is the POSITION of the single survivor, or -1 when the
	// game ended without one (turn budget, wall clock, stall, or a
	// genuine multi-survivor end).
	Winner  int           `json:"winner"`
	Elapsed time.Duration `json:"elapsed_ns"`
	// Stalled is true when the table stopped committing moves. It is
	// REPORTED, not fatal: a stalled game still played twenty turns
	// of real decisions, and throwing those away would lose the
	// evidence that says why it stalled.
	Stalled   bool   `json:"stalled,omitempty"`
	StallDump string `json:"stall_dump,omitempty"`
	// Aborted is true when the RUN's context was cancelled while this
	// game was still being played — a Ctrl-C, not a stall and not a
	// result. The game stopped wherever the cancel found it, so its
	// tallies are a fraction of a game; Run DROPS such a result
	// rather than folding it into every policy's win rate. Never
	// confuse it with Stalled: a stall is evidence about the engine,
	// this is an interruption of the measurement.
	Aborted bool `json:"aborted,omitempty"`
	// DecisionLog and Replay are where this game's artifacts landed.
	DecisionLog string `json:"decision_log,omitempty"`
	Replay      string `json:"replay,omitempty"`
	// DecisionLogStats is the writer's own tally, read after the log
	// was closed and drained. It is reported rather than assumed
	// because the writer is asynchronous and drops rather than
	// blocks a bot seat: a run whose corpus has holes in it should
	// say so on the face of the report, not in a log line nobody
	// kept.
	DecisionLogStats *decisionlog.Stats `json:"decision_log_stats,omitempty"`
}

// WinnerLabel is the tally key of the winning spec, or "" for none.
func (r GameResult) WinnerLabel() string {
	for _, s := range r.Seats {
		if s.Position == r.Winner {
			return s.Spec.Label()
		}
	}
	return ""
}

// samples are one seat's raw measurements for one game.
type samples struct {
	mu          sync.Mutex
	decision    []time.Duration
	modelCall   []time.Duration
	promptBytes []int
}

// Observe implements aiseat.DecisionObserver. It runs inline on the
// runner's goroutine, so it does nothing but append.
func (s *samples) Observe(ev aiseat.DecisionEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.decision = append(s.decision, ev.Latency)
	if ev.Trace.ModelLatency > 0 {
		s.modelCall = append(s.modelCall, ev.Trace.ModelLatency)
	}
	if p := ev.Trace.Prompt; p != nil {
		n := len(p.User)
		for _, b := range p.System {
			n += len(b)
		}
		s.promptBytes = append(s.promptBytes, n)
	}
}

// fanOut sends every event to each of several observers. The arena
// always has at least one of its own (the latency samples) and
// usually a second (the decision log), and Config carries a single
// Observer slot.
type fanOut []aiseat.DecisionObserver

func (f fanOut) Observe(ev aiseat.DecisionEvent) {
	for _, o := range f {
		o.Observe(ev)
	}
}

// Play plays one game and returns what happened.
//
// It is aiseat's playCatalogGame with the stall RETURNED rather than
// fatal, the policies built from tiers rather than hard-coded, and
// the per-seat telemetry kept. order is the seating (see Order); nil
// seats the specs in their configured order.
//
// A stall is not an error. An error is a game that could not be
// STARTED — an unknown deck, a refused player — because that is a
// configuration mistake and playing on would measure nothing.
func Play(ctx context.Context, cfg Config, seed uint64, order []int) (GameResult, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return GameResult{}, err
	}
	n := len(cfg.Seats)
	if order == nil {
		order = Order(n, 0, false)
	}
	if err := checkOrder(order, n); err != nil {
		return GameResult{}, err
	}

	g := game.NewGame()
	for pos := 0; pos < n; pos++ {
		spec := cfg.Seats[order[pos]]
		list, err := seatDeck(cfg, spec)
		if err != nil {
			return GameResult{}, err
		}
		name := fmt.Sprintf("%s%d", spec.Label(), pos)
		if _, err := g.AddPlayer(name, list); err != nil {
			return GameResult{}, fmt.Errorf("botarena: add player %s: %w", name, err)
		}
	}
	// The same two-stream seeding the whole-game tests use, so an
	// arena game and a test game on the same seed deal the same
	// cards.
	if err := g.StartWithFirstPlayerRoll(rand.New(rand.NewPCG(seed, seed+1))); err != nil {
		return GameResult{}, fmt.Errorf("botarena: start: %w", err)
	}

	room := ws.NewRoom(g, cfg.Log, cfg.ReplayDir)
	res := GameResult{Seed: seed, GameID: g.ID, Order: append([]int(nil), order...), Winner: -1}
	if cfg.ReplayDir != "" {
		res.Replay = filepath.Join(cfg.ReplayDir, "replays", g.ID.String()+".jsonl")
	}

	var gameLog *decisionlog.GameLog
	if cfg.DecisionLog != nil {
		gl, err := cfg.DecisionLog.OpenGame(g.ID)
		if err != nil {
			return GameResult{}, fmt.Errorf("botarena: decision log: %w", err)
		}
		gameLog, res.DecisionLog = gl, gl.Path()
	}

	// The wall deadline gets its OWN cause, so that watchTable can
	// tell "my own deadline fired" from "the operator pressed
	// Ctrl-C". Both are ctx.Err() != nil, and reporting the second as
	// the first writes a stall dump claiming a cause that did not
	// happen — into the file an operator files stall reports from.
	ctx, cancel := context.WithTimeoutCause(ctx, cfg.Wall, errWallClock)
	defer cancel()

	started := time.Now()
	runners := make([]*aiseat.Runner, 0, n)
	seats := make([]SeatResult, 0, n)
	metersAndFunnels := make([]seatInstruments, 0, n)
	for pos, p := range g.Seats {
		spec := cfg.Seats[order[pos]]
		pol, inst, err := newSeat(cfg, spec, seed, pos)
		if err != nil {
			if gameLog != nil {
				_ = gameLog.Close()
			}
			return GameResult{}, err
		}
		raw := &samples{}
		rc := cfg.Runner
		if rc.MaxThink <= 0 {
			rc.MaxThink = spec.Tier.RunnerConfigWith(cfg.MaxThink).MaxThink
		}
		// Built by appending rather than as a literal: a nil
		// *decisionlog.GameLog in an interface slot is a non-nil
		// interface holding a nil pointer, and calling Observe on it
		// panics on the runner's own goroutine.
		obs := fanOut{raw}
		if gameLog != nil {
			obs = append(obs, gameLog)
		}
		// Guarded the way gameLog is, and for the same reason: a
		// caller that left a nil *decisionlog.GameLog (or any other
		// nil pointer) in the interface slot hands us a non-nil
		// interface holding a nil pointer, and Observe on it panics
		// on the runner's own goroutine, mid-game.
		if !isNilObserver(cfg.Runner.Observer) {
			obs = append(obs, cfg.Runner.Observer)
		}
		rc.Observer = obs
		runners = append(runners, aiseat.Start(ctx, room, p.ID, pol, rc, nil, cfg.Log))
		seats = append(seats, SeatResult{Spec: spec, Position: pos, Policy: pol.Name(), raw: raw})
		metersAndFunnels = append(metersAndFunnels, inst)
	}

	watchTable(ctx, room, g, cfg, &res)

	cancel()
	for _, r := range runners {
		<-r.Done()
	}
	// Closed only after every runner is done: a runner still in a
	// window would write into a closed file. Close drains the
	// writer, so the stats read after it are final.
	if gameLog != nil {
		// #735's one record per game, written here for the same
		// reason the server's Manager writes it: every seat has
		// exited, so the numbers are final, and it has to land
		// before the file is closed.
		gameLog.ObserveSpend(aiseat.SpendOfRunners(g.ID, runners))
		if err := gameLog.Close(); err != nil {
			cfg.Log.Error("botarena: closing the decision log", "err", err)
		}
		st := gameLog.Stats()
		res.DecisionLogStats = &st
	}

	snap := g.Snapshot()
	res.State, res.Turns, res.Elapsed = snap.State, snap.Turn.Round, time.Since(started)
	g.ReadSnapshot(func() {
		live, alive := -1, 0
		for i, p := range g.Seats {
			seats[i].Life, seats[i].Eliminated = p.Life, p.Eliminated
			if !p.Eliminated {
				live, alive = i, alive+1
			}
		}
		if alive == 1 {
			res.Winner = live
		}
	})
	for i := range seats {
		seats[i].Runner = runners[i].Stats()
		seats[i].Meter = metersAndFunnels[i].meter.Stats()
		if f := metersAndFunnels[i].funnel; f != nil {
			s := f.Stats()
			seats[i].Funnel = &s
		}
		raw := seats[i].raw
		raw.mu.Lock()
		seats[i].Decision = aiseat.PercentilesOf(raw.decision)
		seats[i].ModelCall = aiseat.PercentilesOf(raw.modelCall)
		seats[i].PromptBytesP50 = medianInt(raw.promptBytes)
		raw.mu.Unlock()
	}
	res.Seats = seats
	return res, nil
}

// isNilObserver reports whether an observer slot is empty — nil, or a
// non-nil interface holding a nil pointer. The second is the trap:
// `var gl *decisionlog.GameLog; cfg.Runner.Observer = gl` compares
// != nil and panics on first use.
func isNilObserver(o aiseat.DecisionObserver) bool {
	if o == nil {
		return true
	}
	switch rv := reflect.ValueOf(o); rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer, reflect.Interface:
		return rv.IsNil()
	default:
		return false
	}
}

// seatInstruments are the per-seat measurement handles Play keeps
// hold of: they have to be created before the policy and read after
// the game.
type seatInstruments struct {
	meter  *rules.Meter
	funnel *model.Policy
}

// newSeat builds one seat's policy and the instruments that will be
// read off it.
//
// The meter is PER SEAT, not per table. ADR 0052's report block wants
// the Layer A / B / C share per policy, and for a heuristic seat the
// meter is the only place that share exists (a heuristic seat has no
// model.Stats). One meter shared by a mixed table would add an
// `assisted` seat's absorption to a `heuristic` seat's and make both
// numbers unreadable; the table-wide figure is the sum of these, so
// nothing is lost by splitting them.
func newSeat(cfg Config, spec SeatSpec, seed uint64, pos int) (aiseat.Policy, seatInstruments, error) {
	inst := seatInstruments{meter: &rules.Meter{}}
	opt := tiers.Options{
		Meter:    inst.meter,
		Client:   cfg.Client,
		Models:   cfg.Models,
		MaxThink: cfg.MaxThink,
		Config:   cfg.ModelConfig,
		// Seeded from the game seed and the CHAIR, so that a run is
		// reproducible and two random seats at one table are not
		// playing the same moves.
		Rand: rand.NewPCG(seed, uint64(pos)+1),
	}
	if spec.Tier.NeedsModel() && spec.Deck != "" {
		// The static, prompt-cached half of a model seat's prompt is
		// the decklist it is holding. A model seat given the wrong
		// deck profile is being asked to reason about a list it was
		// never dealt, which is a worse measurement than no profile
		// at all.
		if p, ok := deckprofile.Build(cfg.Index, spec.Deck); ok {
			opt.Deck = p
		}
	}
	pol, err := tiers.New(spec.Tier, opt)
	if err != nil {
		return nil, inst, fmt.Errorf("botarena: %w", err)
	}
	if mp, ok := pol.(*model.Policy); ok {
		inst.funnel = mp
	}
	return pol, inst, nil
}

func seatDeck(cfg Config, spec SeatSpec) ([]game.Card, error) {
	if spec.Deck == "" {
		return BattleDeck(uuid.Nil), nil
	}
	return CuratedDeck(cfg.Index, spec.Deck)
}

func checkOrder(order []int, n int) error {
	if len(order) != n {
		return fmt.Errorf("botarena: order has %d entries for %d seats", len(order), n)
	}
	seen := make([]bool, n)
	for _, k := range order {
		if k < 0 || k >= n || seen[k] {
			return fmt.Errorf("botarena: order %v is not a permutation of 0..%d", order, n-1)
		}
		seen[k] = true
	}
	return nil
}

// errWallClock is the cause this game's own wall deadline cancels
// with. It is what tells a cancelled game apart from an exhausted
// one — see watchTable.
var errWallClock = errors.New("botarena: per-game wall clock exhausted")

// watchInterval is how often the supervision loop looks at the table.
//
// It used to be 2ms. Every pass takes the game's read lock (Snapshot)
// and the room's mutex (Seq) — the same mutex ws.Room.apply holds
// across a Game clone and a dump write — so 500 Hz is ~900k lock
// acquisitions over a 30-minute budget, contending with precisely the
// decision latency this harness reports, and burning a core per game
// to do it. The stall threshold is 3×MaxThink+15s, so polling at
// 100ms costs the stall detector nothing measurable.
const watchInterval = 100 * time.Millisecond

// watchTable is the supervision loop: it stops the game at the turn
// budget, the wall clock or a stall, and fills the evidence in.
func watchTable(ctx context.Context, room *ws.Room, g *game.Game, cfg Config, res *GameResult) {
	lastSeq, lastMove := room.Seq(), time.Now()
	tick := time.NewTicker(watchInterval)
	defer tick.Stop()
	for {
		snap := g.Snapshot()
		if snap.State != game.StateActive || snap.Turn.Round > cfg.TurnBudget {
			return
		}
		if seq := room.Seq(); seq != lastSeq {
			lastSeq, lastMove = seq, time.Now()
		} else if time.Since(lastMove) > cfg.Stall {
			res.Stalled = true
			res.StallDump = stallDump(g, snap.Turn, cfg.Stall)
			return
		}
		if ctx.Err() != nil {
			if errors.Is(context.Cause(ctx), errWallClock) {
				res.Stalled = true
				res.StallDump = fmt.Sprintf("wall clock (%s) exhausted at turn %d", cfg.Wall, snap.Turn.Round)
				return
			}
			// The RUN was cancelled — a Ctrl-C, or a caller's own
			// deadline. This game neither stalled nor finished: it is
			// a fragment, and the only honest thing to say about it
			// is that it was interrupted.
			res.Aborted = true
			return
		}
		select {
		case <-ctx.Done():
		case <-tick.C:
		}
	}
}

// stallDump is everything a reader needs to file the stall: where the
// table froze, what it was waiting for, and every seat's legal moves.
func stallDump(g *game.Game, turn game.Turn, after time.Duration) string {
	var pending int
	var kinds []string
	g.ReadSnapshot(func() {
		pending = len(g.PendingChoices)
		seen := map[game.PendingChoiceKind]bool{}
		for _, c := range g.PendingChoices {
			if c == nil || seen[c.Kind] {
				continue
			}
			seen[c.Kind] = true
			kinds = append(kinds, string(c.Kind))
		}
	})
	sort.Strings(kinds)
	return fmt.Sprintf("no move for %s at turn %d step %s priority=%v pending=%d kinds=[%s]\n%s",
		after, turn.Round, turn.Step, turn.PriorityHolder,
		pending, strings.Join(kinds, ", "), describeSeats(g))
}

// describeSeats prints every seat and every move on offer to it — the
// same dump the whole-game tests print on a stall, because the first
// question about a frozen table is always "what could anybody have
// done".
func describeSeats(g *game.Game) string {
	var b strings.Builder
	type line struct {
		id         uuid.UUID
		name       string
		life, hand int
		elim       bool
	}
	var seats []line
	g.ReadSnapshot(func() {
		for _, p := range g.Seats {
			seats = append(seats, line{p.ID, p.Name, p.Life, p.Hand.Size(), p.Eliminated})
		}
	})
	for _, p := range seats {
		moves := legal.EnumerateFor(g, p.id)
		fmt.Fprintf(&b, "  %s life=%d hand=%d elim=%v moves=%d\n", p.name, p.life, p.hand, p.elim, len(moves))
		for _, m := range moves {
			fmt.Fprintf(&b, "      %s  %s\n", m.Label, string(m.Params))
		}
	}
	g.ReadSnapshot(func() {
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			fmt.Fprintf(&b, "  pending %s chooser=%s reason=%q\n", c.Kind, c.Chooser, c.Reason)
		}
	})
	return b.String()
}

// Run plays Config.Games games and aggregates them.
//
// sink, when non-nil, is called with each GameResult as it finishes,
// on Run's own goroutine. That is how `boteval arena` streams
// games.jsonl: a run of ten model games takes an hour, and a harness
// that writes nothing until the end is a harness whose evidence a
// Ctrl-C destroys.
func Run(ctx context.Context, cfg Config, sink func(GameResult)) (Summary, error) {
	cfg = cfg.withDefaults()
	if err := cfg.Validate(); err != nil {
		return Summary{}, err
	}
	sum := Summary{Config: describeConfig(cfg), Started: time.Now(), PerPolicy: map[string]*PolicyTotals{}}
	acc := newAccumulator(len(cfg.Seats))
	started := time.Now()
	for i := 0; i < cfg.Games; i++ {
		if err := ctx.Err(); err != nil {
			// A cancelled run reports what it managed, rather than
			// throwing away an hour of games because the last one
			// was interrupted.
			finishSummary(&sum, acc, cfg, started)
			return sum, err
		}
		res, err := Play(ctx, cfg, cfg.Seed+uint64(i), Order(len(cfg.Seats), i, cfg.Rotate))
		if err != nil {
			// Finalised exactly like the cancellation path above. The
			// games that DID play can be hours of evidence, and
			// returning a Summary whose PerPolicy is the empty map it
			// started as would print a report with empty Play, Funnel
			// and Latency tables and `elapsed 0s` over a games.jsonl
			// full of real games.
			finishSummary(&sum, acc, cfg, started)
			return sum, err
		}
		if res.Aborted {
			// Cancelled mid-game: a fraction of a game, not a result.
			// Dropped rather than tallied — a half-played game folded
			// into every policy's win rate is worse than no game.
			finishSummary(&sum, acc, cfg, started)
			if cerr := ctx.Err(); cerr != nil {
				return sum, cerr
			}
			return sum, context.Canceled
		}
		if sink != nil {
			sink(res)
		}
		acc.add(res)
		if res.DecisionLogStats != nil {
			sum.DecisionLog.Add(*res.DecisionLogStats)
		}
		sum.Games = append(sum.Games, digest(res))
	}
	finishSummary(&sum, acc, cfg, started)
	return sum, nil
}

// finishSummary derives result-shaped metadata from the games that actually
// completed. Config.Games remains the requested run length, while Chairs and
// ChairWarning describe the evidence present in this summary. That distinction
// matters when cancellation or a later game-start error returns a partial run.
func finishSummary(sum *Summary, acc *accumulator, cfg Config, started time.Time) {
	sum.Elapsed = time.Since(started)
	sum.PerPolicy = acc.totals(len(cfg.Seats))
	played := len(sum.Games)
	sum.Config.Chairs = ChairCounts(len(cfg.Seats), played, cfg.Rotate)
	sum.Config.ChairWarning = ChairBalanceWarning(len(cfg.Seats), played, cfg.Rotate)
}
