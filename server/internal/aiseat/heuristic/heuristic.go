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

	// LeftoverWindows turns on ADR 0126 §5's two windows: the bot's
	// own last main phase and the end step of the seat whose turn
	// comes just before the bot's, each with an empty stack. Mana
	// empties between steps and a tapped permanent untaps in its
	// controller's untap step, so in those windows a move that costs
	// only mana and taps spends nothing the bot would otherwise keep,
	// and it needs to clear only LeftoverThreshold. Off (the zero
	// value) is the pre-S66 heuristic.
	LeftoverWindows bool
	// LeftoverThreshold is the bar a mana-and-taps move clears in one
	// of the two windows (ADR 0126 §5), in place of PassThreshold or
	// InstantThreshold. A move that also costs life, a sacrifice, a
	// discard, another card or a counter keeps the normal bar. Read
	// only while LeftoverWindows is on, and only when it is LOWER than
	// the normal bar.
	LeftoverThreshold float64
	// SpellFloor is the least an untargeted instant or sorcery is worth
	// once it resolves (ADR 0126 §5): a spell whose effect the wire does
	// not carry is priced as a card that replaces itself and does a
	// little more. Just above Weights.Hand, so the cast clears
	// LeftoverThreshold and stays below PassThreshold: cheap spells fill
	// the leftover windows and never crowd out development in the first
	// main phase. A targeted spell is priced by its targets instead
	// (InstantThreshold's "hold it" is unchanged outside the windows,
	// ADR 0126 Out of scope), and an unimplemented card gets no floor.
	// Zero is off, the pre-S66 heuristic.
	SpellFloor float64
	// TapByTiming prices tapping one of the bot's untapped creatures by
	// when it happens (ADR 0126 §5): nothing in the end step just before
	// the bot's turn, because the creature untaps before any opponent
	// attacks; the blocker plus the attack it gives up in the bot's own
	// first main phase, for a creature that could attack, priced as
	// station already prices it; and the flat blocker price elsewhere.
	// Off (the zero value) is the pre-S66 flat price everywhere.
	TapByTiming bool

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

	// SacrificeDyingAnyway prices a permanent sacrificed to pay a cost
	// by the chance the bot would have kept it (ADR 0126 §7,
	// sacrifice.go): a target of an opponent's spell or ability on the
	// stack is kept with 1 − RemovalConfidence, one a declared sweep on
	// the stack would remove is not kept, and so is a creature losing
	// its combat once blockers are declared. A move whose only non-mana
	// cost is such a sacrifice clears LeftoverThreshold, because it
	// spends nothing the bot would otherwise keep. Off (the zero value)
	// charges the whole board value, the pre-S66 price.
	SacrificeDyingAnyway bool
	// DeathPayoff is what each `death_payoff` row on a permanent the bot
	// controls (Blood Artist, Zulaport Cutthroat, Bastion of
	// Remembrance) adds to every creature it sacrifices (ADR 0126 §7).
	// Zero is the pre-S66 price.
	DeathPayoff float64

	// RampPerMana is the cast-time premium per mana a new repeatable
	// mana source closes of the bot's mana deficit (ADR 0126 §2,
	// rampPremium): large while the bot cannot cast what it holds, and
	// nothing once it can, so the premium fades as the game goes on
	// without a turn counter. Zero is the pre-S66 price, and
	// BaselineConfig zeroes it.
	RampPerMana float64
	// RampWantCap caps the mana the deficit aims at: past seven mana,
	// one more source is not what stands between a Commander deck and
	// its hand.
	RampWantCap int

	// PricePurposes turns on ADR 0126 §6's prices (purpose.go): a
	// spell, mode, alternative cost or own activated row that declares
	// what it does is priced by that, in place of the mana-value proxy
	// or ActivateBase, and a permanent's declared enters effect is added
	// to its body. Off (the zero value) is the pre-S66 heuristic.
	PricePurposes bool
	// NetLandSwaps prices a cast that sacrifices the bot's own lands
	// and puts more lands onto the battlefield than it sacrifices
	// (Harrow) by what it nets (#2469, land_swap.go): the lands it
	// sacrifices are replaced one for one, untapped, and only the
	// lands it adds are ramp, with SpellFloor under the net rather
	// than under the gross. Off (the zero value) prices every land the
	// purpose declares as ramp and floors before the sacrifice.
	NetLandSwaps bool
	// TutorWeight is a card searched out to hand or the top of the
	// library, in cards drawn: above one, because the bot picks it.
	TutorWeight float64
	// SelfMillWeight is a card searched out into the graveyard (Entomb),
	// in cards drawn: below one, because only a graveyard plan the
	// policy cannot see makes it a card.
	SelfMillWeight float64
	// DiscardWeight is one card discarded on resolution (a loot's
	// second half). Below Weights.Hand: the bot discards its worst card.
	DiscardWeight float64
	// TokenWeight is one token the purpose makes (a Treasure, a Clue).
	TokenWeight float64
	// AwakenLandShare is the share of a hasty N/N creature an awaken
	// cast's land is worth (ADR 0135 §3, owner decision 6): below one,
	// because the body is a land, and creature removal that answers it
	// also takes a mana source. Zero (the baseline) prices awaken at
	// nothing, so the bot casts an awaken spell for its mana cost.
	AwakenLandShare float64
	// PriceSweeps turns on ADR 0126 §4: a declared sweep is priced as
	// the change in ScoreEval with the permanents it removes taken off
	// the board (sweepValue), so the bot stops casting a wipe onto its
	// own winning board. Off (the zero value) is the pre-S66 proxy.
	PriceSweeps bool
	// DiscardCostByCard turns on the discard half of ADR 0126 §7: a
	// card discarded to pay a spell's additional cost costs what it is
	// worth to the bot (cardValue), not a flat Weights.Hand. Off (the
	// zero value) is the pre-S66 flat price.
	DiscardCostByCard bool
	// LastLandDiscard is what discarding the last land in hand costs on
	// top of its cardValue while the bot is short of LandsWanted: the
	// land drop it may miss (discardCost).
	LastLandDiscard float64
	// PriceDiscardPayoffs turns on ADR 0126's amendment of 2026-10-06:
	// discarding a card that one of the bot's own triggered rows
	// declares a discard_payoff for (Mary Read's Treasure for an
	// Island, Marauding Mako's counter for any card) costs what that
	// payoff pays less (discard_payoff.go). Off (the zero value) prices
	// a discard by the card alone.
	PriceDiscardPayoffs bool
	// PriceExert turns on ADR 0130 §9 (owner decision 4): the twin
	// attack move that exerts its attacker is priced by what the
	// exert's rows declare against what the creature gives up by
	// staying tapped (exert.go), and taken when it is worth more than
	// the plain attack. Off (the zero value) never exerts.
	PriceExert bool
	// ExertCostWeight scales an exert's cost: the creature's next-turn
	// attack value plus its blocking value across the opponents' turns.
	ExertCostWeight float64

	// PlanTurnMana turns on ADR 0136's turn plan: in its own main phase
	// with an empty stack, the bot picks the set of casts this turn's
	// mana buys the most with, and makes that set's first move. Off (the
	// zero value, and BaselineConfig) decides one move at a time.
	//
	// Nothing reads it yet: the plan lands in ADR 0136 PR 4. It exists
	// now so the arena's `heuristic-noplan` contestant (DefaultConfig
	// with it off) is in place for PR 2's baseline, and plays exactly as
	// `heuristic` until then.
	PlanTurnMana bool

	// FuelFloor is what a LAND in a graveyard or in exile is worth to
	// its owner (#1013, fuel.go). The bottom of the scale: a land card
	// in a graveyard does nothing at all without a Crucible, which is
	// why it is the first thing an escape eats. Small and positive,
	// not zero — a floor of zero would make it free rather than
	// cheapest, and free is what "eat it before anything else" already
	// means.
	FuelFloor float64
	// FuelIdle is what any OTHER card in a graveyard or exile is worth
	// when the seat cannot cast it from there. Above FuelFloor, and
	// that gap is the whole of "eat the lands, not the spells": it
	// stands for the graveyard synergies the policy cannot see —
	// delve, a flashback granted later, a Snapcaster target — which
	// want a spell far more often than they want a land.
	FuelIdle float64
	// FuelRecast discounts a graveyard or exile card the seat CAN
	// still cast — escape, flashback, a granted impulse — against the
	// same card in hand. Below one, and that gap is the whole
	// behaviour: an escaping Uro eats the lands rather than the
	// Snapcaster target, and it eats a second Uro last of all. It is a
	// real card and it is not a card in hand: it needs its own cost,
	// its own window, and it can be exiled out from under the plan.
	FuelRecast float64
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

	// SpecialActionValue prices a CR 116.2 special action —
	// foretelling a card, suspending one. Positive and modest: both
	// keywords trade this turn's mana for a cheaper or free cast
	// later, which is real value a seat should take when it has
	// nothing better to do with the mana, and never a reason to skip
	// casting the spell outright (a cast is priced by what it does,
	// and is usually worth more). It is the one price that makes
	// Lotus Bloom and Ancestral Vision playable at all: a card with
	// no mana cost can never be cast from hand (CR 118.6), so
	// suspending it is the only move it will ever have.
	SpecialActionValue float64

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

	// DredgePlaySoon is what a dredged card is worth on top of its
	// cardValue when the bot could play it this turn or next (#2390,
	// dredge.go). A dredge trades an unknown card for a known one, and
	// a known card the bot can play at once is the trade at its best.
	DredgePlaySoon float64
	// DredgeLibraryFloor is the fewest cards a dredge may leave in the
	// bot's library. A dredge that would mill it below this is
	// declined whatever it returns: mill N with a few cards left puts
	// the CR 704.5b decking loss on a timer.
	DredgeLibraryFloor int

	// DevourCommander is what eating the bot's own commander costs on
	// top of its board value (#2419, devour.go): a commander is the
	// win condition and a recurring cast, never fodder.
	DevourCommander float64
	// DevourPermanent is the flat premium on eating a nontoken
	// creature, over and above its stats: a card on the board is an
	// ability the stat line does not show, where a token is just a body.
	DevourPermanent float64

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

	// AttackTaxPenalty is what one point of CR 508.1a attack tax costs
	// an attack's value (ADR 0080, #1063) — Propaganda's {2} is worth
	// 2 × this. It is a MANA price rather than a life or card price,
	// so it is tuned against DamageToOpponent: at the default, {2}
	// cancels roughly a 6-power swing, which is the right shape for
	// "a 1/1 into Ghostly Prison is not worth the turn's mana, a Wurm
	// usually is".
	//
	// Deliberately a flat generic count rather than a colour-aware
	// price: what else the mana could have bought is the fuel pricer's
	// question, not a combat term's, and LethalBonus dwarfs this
	// anyway so a lethal swing still happens at any price the seat can
	// pay.
	AttackTaxPenalty float64

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

		LeftoverWindows:   true,
		LeftoverThreshold: 0.00,
		SpellFloor:        1.30,
		TapByTiming:       true,

		LandValue:      8.00,
		SpellPerMana:   0.60,
		CommanderBonus: 1.50,
		ActivateBase:   0.50,
		RampPerMana:    1.00,
		RampWantCap:    7,

		PricePurposes:     true,
		NetLandSwaps:      true,
		TutorWeight:       1.00,
		SelfMillWeight:    0.50,
		DiscardWeight:     0.60,
		TokenWeight:       0.50,
		AwakenLandShare:   0.75,
		PriceSweeps:       true,
		DiscardCostByCard: true,
		LastLandDiscard:   1.00,

		PriceDiscardPayoffs: true,

		PriceExert:      true,
		ExertCostWeight: 1.00,

		PlanTurnMana: true,

		FuelFloor:  0.05,
		FuelIdle:   0.30,
		FuelRecast: 0.55,
		LifePayoff: 0.35,
		LifeFloor:  1,
		ManaFloat:  -0.50,

		SpecialActionValue: 1.00,

		RemovalConfidence:  0.80,
		LeaderBoost:        1.50,
		DamageToPlayer:     1.20,
		SelfTargetPenalty:  1.50,
		OwnPermanentTarget: 0.40,
		CounterValue:       3.00,
		ScryKeep:           1.00,

		DredgePlaySoon:     0.50,
		DredgeLibraryFloor: 10,
		DevourCommander:    10.0,
		DevourPermanent:    0.25,

		SacrificeDyingAnyway: true,
		DeathPayoff:          0.60,

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
		// One point of tax ≈ one point of power through: at
		// DamageToOpponent 0.30 a {2} tax cancels a 2-power attacker
		// outright and leaves a 7/7 trampler comfortably worth it.
		AttackTaxPenalty: 0.30,

		Concede:      true,
		ConcedeLife:  3,
		ConcedeTurns: 3,
	}
}

// BaselineConfig is the heuristic as it priced cards before S66: the
// frozen reference the arena's `heuristic-baseline` contestant plays
// (ADR 0126 §1 and §9).
//
// ADR 0126 adds its new pricing terms as Config and Weights fields
// whose zero value is the old behaviour. This returns DefaultConfig
// with every one of them zeroed, so a run of `heuristic` against
// `heuristic-baseline` measures exactly what those terms changed.
// Each PR that adds a term zeroes it here in the same change.
//
// TestBaselineConfigRanksTheSuiteAsBefore (aiseat/suite) holds it to
// the rankings the policy gave every suite position before S66,
// whatever DefaultConfig becomes.
func BaselineConfig() Config {
	c := DefaultConfig()
	// §2, mana sources (PR 3).
	c.Weights.ManaPerExtra = 0
	c.RampPerMana = 0
	c.RampWantCap = 0
	// §3, permanents by what they do (PR 4).
	c.Weights.PermanentPerMana = 0
	c.Weights.RowTriggered = 0
	c.Weights.RowStatic = 0
	c.Weights.RowActivated = 0
	c.Weights.RowCap = 0
	// §5, the two windows (PR 5).
	c.LeftoverWindows = false
	c.LeftoverThreshold = 0
	c.SpellFloor = 0
	c.TapByTiming = false
	// §4, §6's prices and §7's discard half (PR 7).
	c.PricePurposes = false
	// #2469: a land swap priced by what it nets.
	c.NetLandSwaps = false
	c.TutorWeight = 0
	c.SelfMillWeight = 0
	c.DiscardWeight = 0
	c.TokenWeight = 0
	// ADR 0135 §3: awaken, priced at nothing before it.
	c.AwakenLandShare = 0
	// ADR 0129 §7: energy, priced at nothing before it.
	c.Weights.Energy = 0
	c.PriceSweeps = false
	c.DiscardCostByCard = false
	c.LastLandDiscard = 0
	// §7, sacrifice outlets (PR 8).
	c.SacrificeDyingAnyway = false
	c.DeathPayoff = 0
	// ADR 0126's amendment of 2026-10-06: discard payoffs.
	c.PriceDiscardPayoffs = false
	// ADR 0130 §9: exert, never taken before it.
	c.PriceExert = false
	c.ExertCostWeight = 0
	// ADR 0136: the turn plan, which the pre-S66 heuristic never had.
	c.PlanTurnMana = false
	return c
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

	// bf, mine, stack, graveyard, exile index the view by instance ID.
	// `mine` is the bot's own hand and command zone — the only
	// hidden zone it is entitled to read.
	bf        map[string]*protocol.CardView
	mine      map[string]*protocol.CardView
	stack     map[string]*protocol.CardView
	graveyard map[string]*protocol.CardView
	// exile is the shared exile pile, which since #673 is a cast
	// surface the bot is offered moves out of (an impulse grant, a
	// foretold card, a warped creature coming back).
	exile   map[string]*protocol.CardView
	choices map[string]*protocol.PendingChoiceView
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
	// beforeMyUntap is the end step of a turn after which the bot's
	// own turn comes next (ADR 0126 §5): the last window before every
	// permanent the bot controls untaps. Stack or no stack.
	beforeMyUntap bool
	// leftover is ADR 0126 §5's window: the bot's own last main phase
	// or beforeMyUntap, with an empty stack. Computed whatever the
	// Config says; Config.LeftoverWindows decides whether it is read.
	leftover bool
	// dying caches dyingAnyway per permanent for this decision (ADR
	// 0126 §7, sacrifice.go).
	dying map[string]float64
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
		turn:      v.Turn.Seq,
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
	for i := range v.Exile.Cards {
		c := &v.Exile.Cards[i]
		if st.exile == nil {
			st.exile = make(map[string]*protocol.CardView, len(v.Exile.Cards))
		}
		st.exile[c.InstanceID] = c
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
	// CR 307.1's empty stack, read off the wire: an activated or
	// triggered ability has no card in stack.cards and shows up only
	// in stack_items (#1352).
	st.sorcerySpeed = st.myTurn && len(v.Stack.Cards) == 0 && len(v.StackItems) == 0 &&
		(st.step == "precombat_main" || st.step == "postcombat_main")
	stackEmpty := len(v.Stack.Cards) == 0 && len(v.StackItems) == 0
	st.beforeMyUntap = st.step == "end" && nextTurnSeat(v) >= 0 && v.Seats[nextTurnSeat(v)].ID == st.me
	st.leftover = stackEmpty &&
		(st.beforeMyUntap || (st.myTurn && st.step == "postcombat_main" && !mainPhaseUpcoming(v)))
	return st
}

// nextTurnSeat is the index into v.Seats of the player whose turn comes
// after this one: the first queued extra turn (CR 500.7), or else the
// next seat in turn order that is still in the game. -1 when the view
// cannot say.
func nextTurnSeat(v *protocol.GameView) int {
	if len(v.Turn.ExtraTurns) > 0 {
		if i := v.Turn.ExtraTurns[0]; i >= 0 && i < len(v.Seats) {
			return i
		}
		return -1
	}
	n, as := len(v.Seats), v.Turn.ActiveSeat
	if as < 0 || as >= n {
		return -1
	}
	for k := 1; k <= n; k++ {
		if i := (as + k) % n; !v.Seats[i].Eliminated {
			return i
		}
	}
	return -1
}

// mainPhaseUpcoming reports whether this turn's plan still holds a main
// phase — an extra combat's postcombat main (CR 505.1a) — so the current
// one is not the turn's last sorcery-speed window.
func mainPhaseUpcoming(v *protocol.GameView) bool {
	for _, u := range v.Turn.Upcoming {
		if u.Step == "precombat_main" || u.Step == "postcombat_main" {
			return true
		}
	}
	return false
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
	// ADR 0121 §4: the opening roll. Roll, or take the first turn —
	// the same answer Layer A gives, from the same function.
	if i, why := aiseat.OpeningRollIndex(in); i >= 0 {
		return aiseat.Decision{Index: i, Reason: why}, nil
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
	threshold := p.cfg.PassThreshold
	if !st.sorcerySpeed {
		threshold = p.cfg.InstantThreshold
	}
	leftover := p.cfg.LeftoverWindows && st.leftover && p.cfg.LeftoverThreshold < threshold
	// best is the highest-priced move, the fallback when no pass is on
	// offer. take is the highest-priced move that clears its OWN bar:
	// with every bar equal (no leftover window) the two are one move
	// whenever take exists, which is the pre-S66 rule exactly.
	best, bestVal, bestReason := -1, 0.0, ""
	take, takeVal, takeReason := -1, 0.0, ""
	for i := range moves {
		if i%16 == 0 && ctx.Err() != nil {
			break
		}
		if moves[i].Kind == legal.KindPass || moves[i].Kind == legal.KindFinishBlocks {
			continue
		}
		v, reason := p.valueOf(st, moves[i])
		if best < 0 || v > bestVal {
			best, bestVal, bestReason = i, v, reason
		}
		bar := threshold
		if leftover && v <= threshold && p.leftoverEligible(st, moves[i]) {
			bar = p.cfg.LeftoverThreshold
			reason += ", leftover mana"
		} else if p.cfg.LeftoverThreshold < threshold && v <= threshold && p.dyingAnywayEligible(st, moves[i]) {
			// ADR 0126 §7: a sacrifice of what the bot is about to lose
			// spends nothing it would keep, §5's premise, in any window.
			bar = p.cfg.LeftoverThreshold
			reason += ", dying anyway"
		}
		if v > bar && (take < 0 || v > takeVal) {
			take, takeVal, takeReason = i, v, reason
		}
	}
	passIdx := indexOfKind(moves, legal.KindPass)
	if passIdx < 0 {
		// #1501: a defender declaring blockers holds no priority, and
		// its "nothing worth doing" is finishing the declaration —
		// the pass it stands in for.
		passIdx = indexOfKind(moves, legal.KindFinishBlocks)
	}
	if take >= 0 {
		return aiseat.Decision{Index: take, Reason: fmt.Sprintf("%s (+%.2f)", takeReason, takeVal)}
	}
	if passIdx >= 0 {
		return aiseat.Decision{Index: passIdx, Reason: "nothing worth doing"}
	}
	if best >= 0 && bestVal > 0 {
		return aiseat.Decision{Index: best, Reason: bestReason + " (no pass on offer)"}
	}
	// #1571: no pass, but an answer the enumerator marks always-legal
	// — the attack a CR 508.1d requirement owes while the active
	// player's pass is withheld. This seat holds priority, so a
	// decline would stall the table; take the owed answer.
	if si := aiseat.SafeIndex(moves); si >= 0 {
		return aiseat.Decision{Index: si, Reason: "owed: " + moves[si].Label}
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
