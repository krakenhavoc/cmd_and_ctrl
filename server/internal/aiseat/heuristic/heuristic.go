// Package heuristic is the rule-based aiseat.Policy: ADR 0033's
// Layer B, the `heuristic` tier on its own, and the fallback under
// every model failure the later tiers can produce. It has to stand
// alone, so it does: no network, no model, no randomness, no state
// carried between games.
//
// # The type gate
//
// This package must never import internal/game, and neither must any
// other policy package under aiseat/. That is ADR 0033 §3's hidden-
// information guarantee, and it is structural rather than a matter of
// discipline: a policy that has no handle on the engine physically
// cannot read an opponent's hand. Everything here reads
// protocol.GameView — the seat's own filtered projection, byte-
// identical to what a human client in this seat receives — and
// []legal.Move, the closed list of things the seat may do. The ban is
// enforced by TestPolicyPackagesDoNotImportGame in imports_test.go,
// which walks every package under aiseat/ and fails the build on a
// direct internal/game import.
//
// # What it is for
//
// The bar is Forge's, stated in ADR 0033 and the S31 issue: makes
// legal moves, makes locally-sensible decisions, doesn't deadlock,
// uses removal on threats. It is not tournament strength and is not
// trying to be. Play quality here is capped by catalog coverage, not
// by the policy.
//
// # Shape of a decision
//
//	Decide
//	 ├─ nothing to decide (0 or 1 moves)            → take it
//	 ├─ mulligan window                             → keep on a
//	 │                                                castable hand
//	 ├─ a pending choice is owed                    → choices.go
//	 ├─ blocks are on offer                         → combat.go
//	 ├─ attacks are on offer                        → combat.go
//	 └─ otherwise                                   → price every move
//	                                                  against passing
package heuristic

import (
	"context"
	"fmt"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// Config tunes the policy on top of the evaluation Weights. The zero
// value is not usable; DefaultConfig is, and New uses it.
type Config struct {
	Weights Weights

	// PassThreshold is how much better than passing a move has to
	// be before the bot takes it in its own main phase. Small and
	// positive: the bot should act, but not for nothing.
	PassThreshold float64
	// InstantThreshold is the same bar outside a sorcery-speed
	// window. Higher, because holding an instant is a real option
	// and a bot that fires its removal at the first legal moment is
	// the classic rule-based-AI tell.
	InstantThreshold float64

	// LandValue prices the once-a-turn land drop. Above every
	// ordinary cast on purpose — land first, then spend.
	LandValue float64
	// SpellPerMana is the value-per-mana proxy for a non-permanent
	// spell whose text the wire does not carry.
	SpellPerMana float64
	// CommanderBonus is the extra value of casting the commander:
	// it is the deck's best card and it comes back when it dies.
	CommanderBonus float64
	// ActivateBase is the flat value of using an activated ability.
	ActivateBase float64
	// LifePayoff is the value proxy for one point of life a move's
	// cost charges — the life-cost twin of SpellPerMana, and there
	// for the same reason. No oracle text reaches a policy, so what
	// an ability DOES is unreadable and the only evidence of how
	// much it does is what it asks for. An ability that charges
	// seven life is presumed to buy about seven life's worth.
	//
	// Above Weights.Life on purpose, and that gap is the whole
	// behaviour: while the bot is comfortable, paying life is a
	// small profit, and as the total falls the quadratic danger
	// term in LifeCostValue overwhelms a linear payoff and the same
	// ability stops being worth it. Griselbrand at 40 draws seven;
	// Griselbrand at 12 does not.
	LifePayoff float64
	// LifeFloor is the life total a move's cost may never take the
	// bot below. One: the seat may spend itself to 1 if the
	// arithmetic really says so, and may never spend itself to 0,
	// because 0 is not a bad position — it is the end of the game
	// (CR 704.5a) and no payoff on the wire can be worth it.
	LifeFloor int
	// ManaFloat prices a bare mana-ability activation. Negative:
	// casts auto-tap, so floating mana is waste.
	ManaFloat float64

	// RemovalConfidence discounts the assumption that a spell which
	// may legally target an opponent's permanent is removal. It is
	// not always true (an unrestricted "target creature gets +3/+3"
	// may target theirs), and this is the number that says so.
	RemovalConfidence float64
	// LeaderBoost multiplies a payoff aimed at the table's biggest
	// threat.
	LeaderBoost float64
	// DamageToPlayer prices a target pointed at an opponent.
	DamageToPlayer float64
	// SelfTargetPenalty prices a target pointed at the bot itself.
	SelfTargetPenalty float64
	// OwnPermanentTarget prices a target pointed at the bot's own
	// permanent — presumed a pump or a protection.
	OwnPermanentTarget float64
	// CounterValue prices countering a spell on the stack.
	CounterValue float64
	// ScryKeep is the indifference point for scry: a card the bot
	// values below this is worth bottoming.
	ScryKeep float64

	// DamageToOpponent prices one point of damage the bot DEALS.
	DamageToOpponent float64
	// DesperateDamage replaces DamageValue once a hit would put the
	// bot at or below BlockChumpLife: this is what buys chump blocks.
	DesperateDamage float64
	// FocusBonus is the extra value of attacking the seat the
	// aggression rotation has settled on, so the bot applies
	// sustained pressure instead of poking whoever is momentarily
	// cheapest.
	FocusBonus float64
	// LethalBonus is the value of a line that would finish a seat
	// outright — an alpha strike the defender cannot absorb, or a
	// spell pointed at a player already inside FinishLife.
	LethalBonus float64
	// FinishLife is the life total at or below which the bot treats
	// "point something at that player" as a kill attempt.
	FinishLife int

	// LandsWanted is how many mana sources the bot wants before it
	// stops valuing lands highly (for mulligans, scry and discard).
	LandsWanted int
	// KeepMinLands / KeepMaxLands bound a keepable opening hand.
	KeepMinLands int
	KeepMaxLands int
	// MaxMulligans caps how far the bot will dig. London mulligans
	// cost a card each; three is already a losing hand.
	MaxMulligans int

	// BlockChumpLife is the life total at or below which the bot
	// will chump-block to survive, throwing away creatures it would
	// otherwise keep.
	BlockChumpLife int
	// AttackReserve is how many untapped creatures the bot keeps
	// home as blockers when an opponent's board threatens it, and
	// ReserveLife is the life total below which it starts doing so.
	// Above ReserveLife the bot swings freely: at 40 life the
	// crack-back is not what kills you, and a Commander bot that
	// never commits never wins.
	AttackReserve int
	ReserveLife   int

	// Concede enables the concede heuristic. On by default and
	// deliberately conservative — see concede.go.
	Concede bool
	// ConcedeLife and ConcedeTurns are the trigger: at or below
	// ConcedeLife, with no board and no hand, for ConcedeTurns
	// consecutive turns.
	ConcedeLife  int
	ConcedeTurns int
}

// DefaultConfig is the shipped tuning.
func DefaultConfig() Config {
	return Config{
		Weights: DefaultWeights(),

		PassThreshold:    0.25,
		InstantThreshold: 1.50,

		LandValue:      8.00,
		SpellPerMana:   0.60,
		CommanderBonus: 1.50,
		ActivateBase:   0.50,
		LifePayoff:     0.35,
		LifeFloor:      1,
		ManaFloat:      -0.50,

		RemovalConfidence:  0.80,
		LeaderBoost:        1.50,
		DamageToPlayer:     1.20,
		SelfTargetPenalty:  1.50,
		OwnPermanentTarget: 0.40,
		CounterValue:       3.00,
		ScryKeep:           1.00,

		DamageToOpponent: 0.30,
		DesperateDamage:  2.00,
		FocusBonus:       1.00,
		LethalBonus:      25.00,
		FinishLife:       3,

		LandsWanted:  5,
		KeepMinLands: 2,
		KeepMaxLands: 5,
		MaxMulligans: 2,

		BlockChumpLife: 8,
		AttackReserve:  1,
		ReserveLife:    25,

		Concede:      true,
		ConcedeLife:  3,
		ConcedeTurns: 3,
	}
}

// Policy is the heuristic aiseat.Policy. Construct one per bot seat:
// it carries the aggression rotation and the concede counter, which
// are per-seat, per-game state. Decide is safe to call from one
// goroutine at a time, which is the Policy contract; the mutex is
// there so ShouldConcede and Decide can be called from the runner
// without a race.
type Policy struct {
	cfg Config

	mu sync.Mutex
	// agg is the aggression rotation (threat.go).
	agg aggression
	// hopelessTurns counts consecutive turns the position has been
	// judged lost; hopelessTurn is the turn the last one was
	// counted on, so one turn cannot count twice.
	hopelessTurns int
	hopelessTurn  int
}

// New returns a heuristic policy with the default tuning.
func New() *Policy { return NewWithConfig(DefaultConfig()) }

// NewWithConfig returns a heuristic policy with explicit tuning. A
// zero Weights inside cfg is filled in with DefaultWeights.
func NewWithConfig(cfg Config) *Policy {
	if cfg.Weights == (Weights{}) {
		cfg.Weights = DefaultWeights()
	}
	return &Policy{cfg: cfg}
}

// Name is the tier name (ADR 0033 §6).
func (p *Policy) Name() string { return "heuristic" }

// Config returns the policy's tuning.
func (p *Policy) Config() Config { return p.cfg }

// Reset drops the per-game state. Policies are stateless between
// games by design (ADR 0033: no learning, no opponent modelling);
// this exists for tests that reuse one policy across tables.
func (p *Policy) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.agg.reset()
	p.hopelessTurns, p.hopelessTurn = 0, 0
}

// state is everything one decision needs, computed once. Building it
// is two passes over the battlefield and one over the seats — well
// under a millisecond on a four-player board, which is what keeps
// the policy inside MaxThink with room to spare.
type state struct {
	w     Weights
	me    string
	view  *protocol.GameView
	evals map[string]*SeatEval
	// order is the seat-order list of player IDs: the deterministic
	// iteration order for anything a decision depends on.
	order []string
	// opps is the live opposition, threat-ranked, highest first.
	opps []*SeatEval
	// leader is the highest-threat opponent, or "" when there is
	// none left.
	leader string

	// bf, mine, stack, graveyard index the view by instance ID.
	// `mine` is the bot's own hand and command zone — the only
	// hidden zone it is entitled to read.
	bf        map[string]*protocol.CardView
	mine      map[string]*protocol.CardView
	stack     map[string]*protocol.CardView
	graveyard map[string]*protocol.CardView
	choices   map[string]*protocol.PendingChoiceView
	// attach resolves the battlefield's attachment relation, so that
	// "what is this permanent worth" answers the same way here as it
	// does inside Evaluate (#727).
	attach attachIndex

	seat         *protocol.PlayerView
	myEval       *SeatEval
	myMana       int
	turn         int
	step         string
	myTurn       bool
	sorcerySpeed bool
}

func (p *Policy) newState(in aiseat.Input) *state {
	v := &in.View
	st := &state{
		w:         p.cfg.Weights,
		me:        in.Seat.String(),
		view:      v,
		bf:        make(map[string]*protocol.CardView, len(v.Battlefield.Cards)),
		mine:      map[string]*protocol.CardView{},
		stack:     make(map[string]*protocol.CardView, len(v.Stack.Cards)),
		graveyard: map[string]*protocol.CardView{},
		turn:      v.Turn.Number,
		step:      v.Turn.Step,
	}
	st.evals = st.w.Evaluate(*v)
	st.attach = newAttachIndex(v.Battlefield.Cards)
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		st.bf[c.InstanceID] = c
	}
	for i := range v.Stack.Cards {
		c := &v.Stack.Cards[i]
		st.stack[c.InstanceID] = c
	}
	for i := range v.Seats {
		s := &v.Seats[i]
		st.order = append(st.order, s.ID)
		for j := range s.Graveyard.Cards {
			c := &s.Graveyard.Cards[j]
			st.graveyard[c.InstanceID] = c
		}
		if s.ID != st.me {
			continue
		}
		st.seat = s
		for j := range s.Hand.Cards {
			c := &s.Hand.Cards[j]
			st.mine[c.InstanceID] = c
		}
		for j := range s.Command.Cards {
			c := &s.Command.Cards[j]
			st.mine[c.InstanceID] = c
		}
	}
	for i := range v.PendingChoices {
		c := &v.PendingChoices[i]
		if st.choices == nil {
			st.choices = make(map[string]*protocol.PendingChoiceView, len(v.PendingChoices))
		}
		st.choices[c.ID] = c
	}
	st.myEval = st.evals[st.me]
	if st.myEval != nil {
		st.myMana = st.myEval.UntappedMana
	}
	st.opps = st.w.rankOpponents(st.evals, st.me, st.order)
	if len(st.opps) > 0 {
		st.leader = st.opps[0].ID
	}
	as := v.Turn.ActiveSeat
	st.myTurn = as >= 0 && as < len(v.Seats) && v.Seats[as].ID == st.me
	st.sorcerySpeed = st.myTurn && len(v.Stack.Cards) == 0 &&
		(st.step == "precombat_main" || st.step == "postcombat_main")
	return st
}

// permanentValue is boardValue against this decision's battlefield:
// the one pricing of a permanent the whole policy uses, with an
// attached permanent priced by its role rather than on its own line
// (#727). A card that is not on the battlefield — one in hand, one in
// a graveyard — is attached to nothing and prices exactly as it always
// did.
func (st *state) permanentValue(c *protocol.CardView) float64 {
	return st.w.boardValue(c, st.attach)
}

// Decide is the aiseat.Policy entry point.
func (p *Policy) Decide(ctx context.Context, in aiseat.Input) (aiseat.Decision, error) {
	if len(in.Moves) == 0 {
		return aiseat.Decision{}, aiseat.ErrNoMoves
	}
	if len(in.Moves) == 1 {
		return aiseat.Decision{Index: 0, Reason: "only legal move"}, nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	st := p.newState(in)

	// Windows where the enumerator offers one family of moves and
	// nothing else. Each is answered on its own terms.
	switch {
	case allKind(in.Moves, legal.KindMulligan):
		return p.decideMulligan(st, in.Moves), nil
	case allKind(in.Moves, legal.KindChoice):
		return p.decideChoice(ctx, st, in.Moves), nil
	}

	// Combat first: a block that saves eight damage beats any cast on
	// offer in the same window, and an attack is the only way the bot
	// ever wins.
	if anyKind(in.Moves, legal.KindBlock) {
		if d, ok := p.decideBlock(st, in.Moves); ok {
			return d, nil
		}
	}
	if anyKind(in.Moves, legal.KindAttack) {
		if d, ok := p.decideAttack(st, in.Moves); ok {
			return d, nil
		}
	}

	return p.decideGeneral(ctx, st, in.Moves), nil
}

// decideGeneral prices every move against passing and takes the best
// one that clears the bar. It is the only loop long enough to care
// about ctx, and it returns best-so-far the moment the deadline
// lands rather than blowing through it (ADR 0033 §10 — the table
// never waits on a bot).
func (p *Policy) decideGeneral(ctx context.Context, st *state, moves []legal.Move) aiseat.Decision {
	best, bestVal, bestReason := -1, 0.0, ""
	for i := range moves {
		if i%16 == 0 && ctx.Err() != nil {
			break
		}
		if moves[i].Kind == legal.KindPass {
			continue
		}
		v, reason := p.valueOf(st, moves[i])
		if best < 0 || v > bestVal {
			best, bestVal, bestReason = i, v, reason
		}
	}
	threshold := p.cfg.PassThreshold
	if !st.sorcerySpeed {
		threshold = p.cfg.InstantThreshold
	}
	passIdx := indexOfKind(moves, legal.KindPass)
	if best >= 0 && bestVal > threshold {
		return aiseat.Decision{Index: best, Reason: fmt.Sprintf("%s (+%.2f)", bestReason, bestVal)}
	}
	if passIdx >= 0 {
		return aiseat.Decision{Index: passIdx, Reason: "nothing worth doing"}
	}
	if best >= 0 && bestVal > 0 {
		return aiseat.Decision{Index: best, Reason: bestReason + " (no pass on offer)"}
	}
	// No pass means this seat does not hold priority — a combat
	// declaration window, most likely. Declining is safe there and
	// the runner turns a decline into a pass whenever one exists.
	return aiseat.Decision{Index: aiseat.Decline, Reason: "nothing worth doing"}
}

// decideMulligan keeps any hand that can cast something. Two to five
// lands in seven is the standard keepable range; below the floor the
// hand cannot function and above the ceiling it is all lands.
func (p *Policy) decideMulligan(st *state, moves []legal.Move) aiseat.Decision {
	keep := indexOfType(moves, legal.TypeKeepHand)
	mull := indexOfType(moves, legal.TypeMulligan)
	if keep < 0 {
		return aiseat.Decision{Index: 0, Reason: "no keep on offer"}
	}
	if mull < 0 || st.seat == nil {
		return aiseat.Decision{Index: keep, Reason: "keep (no mulligan on offer)"}
	}
	if st.seat.MulligansTaken >= p.cfg.MaxMulligans {
		return aiseat.Decision{Index: keep, Reason: "keep (out of mulligans)"}
	}
	lands, size := 0, len(st.seat.Hand.Cards)
	for i := range st.seat.Hand.Cards {
		if isLand(&st.seat.Hand.Cards[i]) {
			lands++
		}
	}
	next := decode[mulliganParams](moves[mull].Params).HandSize
	if next <= 5 {
		// Digging below six costs more than a bad hand does.
		return aiseat.Decision{Index: keep, Reason: "keep (not digging past six)"}
	}
	lo, hi := p.cfg.KeepMinLands, p.cfg.KeepMaxLands
	if size > 0 && size < 7 {
		// A smaller hand needs proportionally fewer lands.
		hi = size - 1
	}
	if lands < lo || lands > hi {
		return aiseat.Decision{
			Index:  mull,
			Reason: fmt.Sprintf("mulligan: %d lands in %d", lands, size),
		}
	}
	return aiseat.Decision{Index: keep, Reason: fmt.Sprintf("keep: %d lands in %d", lands, size)}
}

// --- small helpers over a move list -------------------------------

func anyKind(moves []legal.Move, k legal.Kind) bool {
	for i := range moves {
		if moves[i].Kind == k {
			return true
		}
	}
	return false
}

func allKind(moves []legal.Move, k legal.Kind) bool {
	for i := range moves {
		if moves[i].Kind != k {
			return false
		}
	}
	return len(moves) > 0
}

func indexOfKind(moves []legal.Move, k legal.Kind) int {
	for i := range moves {
		if moves[i].Kind == k {
			return i
		}
	}
	return -1
}

func indexOfType(moves []legal.Move, t string) int {
	for i := range moves {
		if moves[i].Type == t {
			return i
		}
	}
	return -1
}
