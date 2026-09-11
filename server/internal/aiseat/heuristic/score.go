package heuristic

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// score.go is the board-state evaluation the whole policy rests on:
// one number per seat for "how well is this seat doing", and one
// number per table for "how well am I doing relative to everyone
// else". Everything it reads comes off protocol.GameView — the same
// filtered projection a human at this seat receives — so the
// evaluation can never see a card the seat is not allowed to see.
// An opponent's hand contributes its COUNT and nothing else, which
// is exactly the information a player at the table has.

// Weights are the evaluation's tunable constants. DefaultWeights is
// what the `heuristic` tier ships with; the struct is exported so a
// future tier (or a test) can perturb one term without forking the
// function. Units are arbitrary but consistent: one point is
// roughly "one point of creature power".
type Weights struct {
	// Life is per point of life. Commander starts at 40, so life is
	// deliberately cheap per point — 40 life is not worth more than
	// a board.
	Life float64
	// LifeDanger is the extra penalty per point of life below
	// DangerLife, applied quadratically. Being at 4 is much worse
	// than twice as bad as being at 8.
	LifeDanger float64
	// DangerLife is where LifeDanger starts to bite.
	DangerLife int
	// Hand is per card in hand — cards are resources.
	Hand float64
	// Library is per card left in the library: decking is a real
	// loss condition, but a distant one.
	Library float64

	// Power and Toughness value a creature's printed-plus-layered
	// stats. Power is worth more than toughness.
	Power     float64
	Toughness float64
	// Keyword scales the keyword table in keywordBonus.
	Keyword float64
	// TappedCreature and SickCreature are multipliers on a creature
	// that cannot block or cannot act yet.
	TappedCreature float64
	SickCreature   float64

	// Permanent is the flat value of a non-creature, non-land
	// permanent. Planeswalker adds on top of it, and Loyalty values
	// each loyalty counter.
	Permanent    float64
	Planeswalker float64
	Loyalty      float64

	// ManaSource is per untapped mana source; TappedManaSource is
	// what a tapped one is still worth (it untaps next turn).
	ManaSource       float64
	TappedManaSource float64

	// CommanderTax is the penalty per commander cast already made —
	// the {2} surcharge compounds and a seat that has recast its
	// commander three times is genuinely worse off.
	CommanderTax float64

	// Unknown is what a permanent whose face this seat cannot read
	// (a face-down creature) is worth. Deliberately small and
	// positive: it is something rather than nothing.
	Unknown float64

	// OpponentMean and OpponentMax weight the opposition term in
	// Score. Both are subtracted, so the max term is what makes
	// removal land on the table's leader rather than on whoever
	// happens to be first in seat order.
	OpponentMean float64
	OpponentMax  float64

	// ThreatLifeRev is per point of life an opponent is MISSING,
	// for threat ranking only: a player at 12 is a more attractive
	// place to point damage than a player at 38 with the same board.
	ThreatLifeRev float64
	// ThreatHand weights an opponent's hand in the threat ranking —
	// unknown cards are scarier than known ones.
	ThreatHand float64
}

// StartingLife is the Commander starting total the threat ranking
// measures "life missing" against.
const StartingLife = 40

// DefaultWeights is the shipped tuning. The numbers are the usual
// rule-based-AI ratios (Forge's are in the same family) rather than
// anything derived: a 2/2 is worth ~2.9, a land ~1.0, a card in hand
// ~1.2, a point of life ~0.12.
func DefaultWeights() Weights {
	return Weights{
		Life:       0.12,
		LifeDanger: 0.20,
		DangerLife: 10,
		Hand:       1.20,
		Library:    0.01,

		Power:          1.00,
		Toughness:      0.45,
		Keyword:        1.00,
		TappedCreature: 0.85,
		SickCreature:   0.90,

		Permanent:    1.20,
		Planeswalker: 3.00,
		Loyalty:      0.40,

		ManaSource:       1.00,
		TappedManaSource: 0.55,

		CommanderTax: 1.00,
		Unknown:      1.50,

		OpponentMean: 1.00,
		OpponentMax:  0.50,

		ThreatLifeRev: 0.10,
		ThreatHand:    0.80,
	}
}

// keywordTable is the per-keyword bonus added to a creature's value,
// before the Keyword weight scales it. Keys are the canonical
// lowercase tokens the engine puts on CardView.Abilities (AGENTS.md
// §7). Anything not in the table is worth nothing, which is the
// right default for the long tail of ability words.
var keywordTable = map[string]float64{
	"flying":            1.50,
	"trample":           1.00,
	"deathtouch":        1.60,
	"lifelink":          1.00,
	"first strike":      1.00,
	"double strike":     2.50,
	"vigilance":         0.60,
	"menace":            0.80,
	"reach":             0.40,
	"haste":             0.50,
	"hexproof":          1.60,
	"shroud":            0.60,
	"indestructible":    2.20,
	"protection":        1.20,
	"ward":              1.00,
	"flash":             0.25,
	"skulk":             0.30,
	"fear":              0.80,
	"intimidate":        0.80,
	"horsemanship":      1.20,
	"defender":          -1.50,
	"decayed":           -0.80,
	"cumulative upkeep": -0.60,
}

// keywordBonus sums the table over a card's effective ability list.
// "protection from …" and "ward {2}" arrive as prefixed strings, so
// the match is a prefix match for those two.
func keywordBonus(c *protocol.CardView) float64 {
	var sum float64
	for _, a := range c.Abilities {
		k := strings.ToLower(strings.TrimSpace(a))
		if v, ok := keywordTable[k]; ok {
			sum += v
			continue
		}
		switch {
		case strings.HasPrefix(k, "protection"):
			sum += keywordTable["protection"]
		case strings.HasPrefix(k, "ward"):
			sum += keywordTable["ward"]
		}
	}
	return sum
}

// SeatEval is the per-seat breakdown Evaluate computes on its way to
// a single number. Exposed because the threat ranking, the attack
// planner and the concede heuristic all want the parts, not the
// total, and recomputing them three times a decision is waste.
type SeatEval struct {
	ID         string
	Seat       int
	Life       int
	Hand       int
	Library    int
	Eliminated bool

	// Creatures is the creature portion of Board; Board is every
	// permanent this seat controls.
	Creatures float64
	Board     float64
	// UntappedMana is the count of untapped permanents that can tap
	// for mana — the bot's read on "can they respond?".
	UntappedMana int
	// CreatureCount and UntappedCreatures are the combat-relevant
	// counts.
	CreatureCount     int
	UntappedCreatures int

	// Strength is the seat's absolute standing: board + life + cards
	// − commander tax.
	Strength float64
}

// isType reports whether a card's effective type line names t.
func isType(c *protocol.CardView, t string) bool {
	return strings.Contains(strings.ToLower(c.TypeLine), t)
}

func isCreature(c *protocol.CardView) bool { return isType(c, "creature") }
func isLand(c *protocol.CardView) bool     { return isType(c, "land") }

// CreatureValue is what one creature on the battlefield is worth.
// Exported because the combat planner trades creatures against each
// other and must use the same scale as the board evaluation.
func (w Weights) CreatureValue(c *protocol.CardView) float64 {
	v := w.Power*float64(c.Power) + w.Toughness*float64(c.Toughness) + w.Keyword*keywordBonus(c)
	if v < 0.25 {
		// Even a 0/1 wall is a body; never let a creature price at
		// zero or the blocker logic stops caring whether it dies.
		v = 0.25
	}
	if c.Tapped {
		v *= w.TappedCreature
	}
	if c.SummoningSick {
		v *= w.SickCreature
	}
	return v
}

// CombatValue is CreatureValue without the tapped and summoning-sick
// discounts. Combat has to compare an attacker — which is tapped the
// moment it is declared — against an untapped blocker, and the board
// evaluation's "a tapped creature is worth less" discount would make
// every attacker look cheap and the bot would never block.
func (w Weights) CombatValue(c *protocol.CardView) float64 {
	v := w.Power*float64(c.Power) + w.Toughness*float64(c.Toughness) + w.Keyword*keywordBonus(c)
	if v < 0.25 {
		v = 0.25
	}
	return v
}

// MarginalLife is what ONE more point of life is worth to a seat at
// the given total: the linear term plus the slope of the low-life
// penalty. At 40 it is small — Commander players do not block to save
// two damage — and it climbs steeply as the total falls, which is
// what turns "never chump" into "chump to survive" without a second
// rule saying so.
func (w Weights) MarginalLife(life int) float64 {
	v := w.Life
	if life < w.DangerLife {
		v += 2 * w.LifeDanger * float64(w.DangerLife-life)
	}
	return v
}

// permanentValue prices one permanent on the battlefield.
func (w Weights) permanentValue(c *protocol.CardView) float64 {
	if !c.KnownByYou && (c.FaceDown || c.Name == "") {
		return w.Unknown
	}
	switch {
	case isCreature(c):
		return w.CreatureValue(c)
	case isType(c, "planeswalker"):
		return w.Permanent + w.Planeswalker + w.Loyalty*float64(c.Counters["loyalty"])
	case len(c.ManaAbilities) > 0 || isLand(c):
		if c.Tapped {
			return w.TappedManaSource
		}
		return w.ManaSource
	default:
		return w.Permanent
	}
}

// producesMana reports whether a permanent can be tapped for mana
// right now.
func producesMana(c *protocol.CardView) bool {
	if c.Tapped {
		return false
	}
	if len(c.ManaAbilities) > 0 {
		return !(isCreature(c) && c.SummoningSick)
	}
	return isLand(c)
}

// Evaluate breaks the view down per seat: one pass over the
// battlefield, one pass over the seats.
func (w Weights) Evaluate(v protocol.GameView) map[string]*SeatEval {
	out := make(map[string]*SeatEval, len(v.Seats))
	for i := range v.Seats {
		s := &v.Seats[i]
		out[s.ID] = &SeatEval{
			ID:         s.ID,
			Seat:       s.Seat,
			Life:       s.Life,
			Hand:       s.Hand.Count,
			Library:    s.Library.Count,
			Eliminated: s.Eliminated,
		}
	}
	for i := range v.Battlefield.Cards {
		c := &v.Battlefield.Cards[i]
		e := out[c.Controller]
		if e == nil {
			continue
		}
		e.Board += w.permanentValue(c)
		if isCreature(c) {
			e.Creatures += w.CreatureValue(c)
			e.CreatureCount++
			if !c.Tapped {
				e.UntappedCreatures++
			}
		}
		if producesMana(c) {
			e.UntappedMana++
		}
	}
	for i := range v.Seats {
		s := &v.Seats[i]
		e := out[s.ID]
		if e.Eliminated {
			e.Strength = eliminatedStrength
			continue
		}
		var tax float64
		for _, n := range s.CommanderCasts {
			tax += float64(n)
		}
		e.Strength = e.Board +
			w.Life*float64(e.Life) +
			w.Hand*float64(e.Hand) +
			w.Library*float64(e.Library) -
			w.CommanderTax*tax -
			w.lifeDanger(e.Life)
	}
	return out
}

// eliminatedStrength is the standing of a seat that is out of the
// game: far below anything a live seat can reach, so "eliminate that
// player" always dominates "shrink their board".
const eliminatedStrength = -1000.0

func (w Weights) lifeDanger(life int) float64 {
	if life >= w.DangerLife {
		return 0
	}
	d := float64(w.DangerLife - life)
	return w.LifeDanger * d * d
}

// Score is the positional evaluation of the whole table from one
// seat's point of view: my standing, less the table's. The mean term
// is what makes "everyone else is bigger than me" a bad position
// even when no single opponent is; the max term is what makes the
// LEADER the right place to point removal.
//
// perspective is a seat's player-UUID string, as it appears in
// PlayerView.ID. An unknown perspective scores 0.
func Score(v protocol.GameView, perspective string) float64 {
	return DefaultWeights().Score(v, perspective)
}

// Score is the Weights-tunable form of the package-level Score.
func (w Weights) Score(v protocol.GameView, perspective string) float64 {
	return w.ScoreEval(w.Evaluate(v), perspective)
}

// ScoreEval is Score over an already-computed breakdown.
func (w Weights) ScoreEval(evals map[string]*SeatEval, perspective string) float64 {
	me := evals[perspective]
	if me == nil {
		return 0
	}
	var sum, max float64
	n := 0
	for id, e := range evals {
		if id == perspective || e.Eliminated {
			continue
		}
		sum += e.Strength
		if n == 0 || e.Strength > max {
			max = e.Strength
		}
		n++
	}
	if n == 0 {
		return me.Strength
	}
	return me.Strength - w.OpponentMean*(sum/float64(n)) - w.OpponentMax*max
}

// Threat ranks an opponent as a place to point damage and removal:
// their board, their hidden resources, and how close they already
// are to dying. It is deliberately NOT Strength — a player at 8 life
// with an empty board is a low-strength seat and a high-priority
// attack target, and the bot has to be able to tell those apart.
func (w Weights) Threat(e *SeatEval) float64 {
	if e == nil || e.Eliminated {
		return eliminatedStrength
	}
	missing := StartingLife - e.Life
	if missing < 0 {
		missing = 0
	}
	return e.Board + w.ThreatHand*float64(e.Hand) + w.ThreatLifeRev*float64(missing)
}
