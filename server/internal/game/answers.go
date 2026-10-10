package game

import "strings"

// answers.go — ADR 0142 decision 1: what an ability can do in response.
//
// Purpose.Answers is a small declared set, a bit set so Purpose stays
// comparable. Each value has a tier, fixed here and never declared:
// a STACK answer can matter against any item on the stack (a
// regeneration shield, a sacrifice outlet, a fog); a COMBAT answer
// changes a fight and counts only in a combat window (a granted flying,
// a manland, a token blocker). AnswerValue is the declared "answers
// nothing", and is refused beside any other value.
//
// Like every Purpose field it is declared by hand on the card file and
// never derived. The engine never reads it. internal/legal reads it to
// set a move's interacts and combat_interacts flags (answersOf), and
// falls back to the printed-text read for a row that declares none.

// Answers is a set of declared answers. The zero value is "not
// declared".
type Answers uint16

const (
	// AnswerProtect keeps a permanent of yours on the battlefield, or
	// out of reach: regenerate, indestructible, hexproof, shroud,
	// protection, phase out, a blink, returning itself to hand, granting
	// persist or undying.
	AnswerProtect Answers = 1 << iota
	// AnswerPump raises power or toughness: +N/+N, +1/+1 counters,
	// monstrosity, adapt, a base power and toughness set until end of
	// turn.
	AnswerPump
	// AnswerPrevent prevents or redirects damage, or sets a damage
	// shield.
	AnswerPrevent
	// AnswerRemove removes, destroys, damages or shrinks other
	// permanents with no target: damage to each creature, destroy all,
	// -N/-N to each, an edict.
	AnswerRemove
	// AnswerSacOutlet lets its controller sacrifice a creature at will,
	// in its cost or its effect.
	AnswerSacOutlet
	// AnswerRestrict stops what an opponent may do next: can't cast,
	// can't activate.
	AnswerRestrict
	// AnswerCombatGrant grants a combat keyword or permission: flying,
	// menace, trample, haste, lifelink, vigilance, reach, first strike,
	// double strike, deathtouch, "can't be blocked", "can block an
	// additional creature".
	AnswerCombatGrant
	// AnswerAnimate becomes a creature until end of turn: a creature
	// land, a Vehicle that animates itself, crew.
	AnswerAnimate
	// AnswerMakesBlocker creates one or more creature tokens at instant
	// speed.
	AnswerMakesBlocker
	// AnswerValue is the declared "answers nothing": draw, mana, ramp, a
	// fetch, scry, a non-creature token, a counter that only counts. It
	// stands alone.
	AnswerValue

	// answersKnown is every bit above.
	answersKnown = AnswerValue<<1 - 1
)

// AnswerTier is when an answer matters.
type AnswerTier int

const (
	// TierNone is AnswerValue's: it matters nowhere.
	TierNone AnswerTier = iota
	// TierStack answers can matter against any item on the stack.
	TierStack
	// TierCombat answers change attacks or blocks, and count only in a
	// combat window.
	TierCombat
)

// answerInfo is one vocabulary entry: its bit, its wire name and its
// tier.
type answerInfo struct {
	Bit  Answers
	Wire string
	Tier AnswerTier
}

// answerTable is the vocabulary in the ADR's order, which is also the
// order the wire lists a set in.
var answerTable = []answerInfo{
	{AnswerProtect, "protect", TierStack},
	{AnswerPump, "pump", TierStack},
	{AnswerPrevent, "prevent", TierStack},
	{AnswerRemove, "remove", TierStack},
	{AnswerSacOutlet, "sac_outlet", TierStack},
	{AnswerRestrict, "restrict", TierStack},
	{AnswerCombatGrant, "combat_grant", TierCombat},
	{AnswerAnimate, "animate", TierCombat},
	{AnswerMakesBlocker, "makes_blocker", TierCombat},
	{AnswerValue, "value", TierNone},
}

// AnswerWireNames lists every wire name, in the vocabulary's order. The
// documentation and the client mirror it.
func AnswerWireNames() []string {
	out := make([]string, len(answerTable))
	for i, e := range answerTable {
		out[i] = e.Wire
	}
	return out
}

// Has reports whether every answer in b is in a.
func (a Answers) Has(b Answers) bool { return b != 0 && a&b == b }

// HasAny reports whether a and b share an answer.
func (a Answers) HasAny(b Answers) bool { return a&b != 0 }

// Unknown is the bits of a that name no answer.
func (a Answers) Unknown() Answers { return a &^ answersKnown }

// HasTier reports whether any answer in the set is of tier t.
func (a Answers) HasTier(t AnswerTier) bool {
	for _, e := range answerTable {
		if a&e.Bit != 0 && e.Tier == t {
			return true
		}
	}
	return false
}

// Wire is the set's wire names, in the vocabulary's order. Nil for the
// empty set.
func (a Answers) Wire() []string {
	var out []string
	for _, e := range answerTable {
		if a&e.Bit != 0 {
			out = append(out, e.Wire)
		}
	}
	return out
}

// String is the wire names joined with "|", "-" for none.
func (a Answers) String() string {
	w := a.Wire()
	if len(w) == 0 {
		return "-"
	}
	return strings.Join(w, "|")
}
