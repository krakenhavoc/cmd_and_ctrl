package game

import (
	"fmt"

	"github.com/google/uuid"
)

// cost_modifier.go — S28: "spells cost {1} more / {2} less to cast"
// (CR 601.2f). The fifth kind of cost the engine models, and the
// only one that belongs to a card OTHER than the one being cast.
//
// The four that came before it are all properties of the spell: its
// mana cost (S15), an activated ability's cost (S21 sub-PR 2), an
// additional cost (S21 sub-PR 5) and an alternative cost (S22). A
// cost modifier is a property of a PERMANENT SOMEONE ELSE CONTROLS,
// consulted every time anybody announces anything. That inversion is
// why it gets its own file and its own gathering pass rather than a
// slot on Spec next to AdditionalCost: the cast path has a card in
// hand and has to go LOOKING for the modifiers, exactly as the layer
// engine goes looking for static abilities.
//
// # Ordering (CR 601.2f)
//
// The rule fixes the order and the order is observable:
//
//	1. the cost to pay — printed mana cost, or the alternative cost
//	   claimed at announce (CR 601.2e, already settled by the time
//	   this file runs);
//	2. additional costs (paid, not priced — see additional_cost.go);
//	3. cost INCREASES, in any order among themselves;
//	4. cost REDUCTIONS, in any order among themselves;
//	5. cost-SETTING effects last (Trinisphere).
//
// Increases before reductions is the half that changes outcomes.
// Sphere of Resistance + Goblin Electromancer on a Lightning Bolt is
// {R} → {1}{R} → {R}: the reduction has a generic symbol to eat
// because the increase already put one there. Reversed, the
// reduction would find nothing to reduce and the Bolt would cost
// {1}{R}. The engine applies them in two passes for exactly this
// reason.
//
// # The floor (CR 601.2f)
//
// A cost reduction reduces GENERIC mana and nothing else. It can
// take the generic component to zero and no further — there is no
// negative mana and no "credit" that carries to the next reduction —
// and it may never touch a coloured requirement. Heartless Summoning
// on a {B} Carrion Feeder leaves {B}, not free and not {0}. This is
// the most commonly mis-implemented rule in the family, and
// reduceGeneric is deliberately the only place in the engine that
// knows how to spend a reduction.
//
// (The sprint brief cited "a floor of {1} per CR 117.13". There is
// no such rule: CR 117 is timing and priority, and the real floor is
// zero generic mana with coloured requirements untouched, which is
// what this file implements. A {1} floor would make Heartless
// Summoning + a {1}{B} creature cost {1}{B}, which is wrong on the
// card.)
//
// # Failure
//
// A modifier that returns a nonsensical amount REFUSES THE CAST. It
// does not clamp, and it does not fall through to the printed cost.
// applyCastCostLocked learned this the hard way in #289, where an
// unreadable cost silently made ~1,047 cards free; a reducer that
// returns a negative number, or an increaser that returns one, is
// the same failure wearing a different hat. Refusing is the only
// answer that can't ship a card cheaper than printed.

// CostModifierKind places a modifier in the CR 601.2f order. The
// zero value is CostIncrease, which is the safe default: a modifier
// that forgot to declare its kind makes spells more expensive rather
// than less.
type CostModifierKind int

const (
	// CostIncrease is "spells cost {N} more to cast" — Sphere of
	// Resistance, Thalia, Thorn of Amethyst, Aura of Silence,
	// Damping Sphere. Amount is the generic mana added.
	CostIncrease CostModifierKind = iota

	// CostReduction is "spells you cast cost {N} less to cast" —
	// Goblin Electromancer, Heartless Summoning, Animar. Amount is
	// the generic mana removed, floored at the cost's generic
	// component; coloured requirements are never touched.
	CostReduction

	// CostFloor is a cost-SETTING effect: "each spell that would
	// cost less than three mana to cast costs three mana to cast"
	// (Trinisphere). Amount is the MINIMUM total mana value, and
	// the shortfall is made up in generic mana.
	//
	// A separate kind rather than a conditional increase because
	// the ordering is separate: Trinisphere's own ruling is that
	// it applies after every other increase and reduction, which
	// is the only way "would cost less than three" can mean the
	// cost the spell actually ends up at.
	CostFloor
)

// CostQuery is everything a cost modifier may look at. Passed by
// value: a modifier is consulted inside the cast path under the
// write lock, and handing it pointers into the battlefield slice
// would invite a card file to mutate the board while pricing a
// spell.
//
// Game is a pointer because the interesting predicates are
// board-wide reads — "for each other spell that player has cast this
// turn" (Damping Sphere), "for each +1/+1 counter on Animar". Treat
// it as READ-ONLY: a *ForEffect accessor is fine, a mutator is a
// bug and a public locking mutator is a deadlock.
type CostQuery struct {
	// Game is the game the cast is happening in. Read-only.
	Game *Game

	// Card is the spell being cast, as it stands at announce. For a
	// multi-face card this is the face that was chosen (ADR 0034),
	// because the cast path stamps the face before pricing.
	Card Card

	// Controller is the player casting the spell — the "you" in
	// "spells YOU cast cost {1} less", and the player "an opponent"
	// is measured against in "spells your opponents cast".
	Controller uuid.UUID

	// Source is the permanent contributing the modifier. Its
	// Controller is the "you" of the ABILITY, as against Controller
	// above which is the "you" of the CAST — Aura of Silence needs
	// both to tell an opponent's artifact spell from its own
	// controller's.
	//
	// Its Tapped bit is live, which is what lets Trinisphere's "as
	// long as this artifact is untapped" be an ordinary predicate
	// rather than special machinery.
	Source Card

	// FromZone is where the spell is being cast from — hand, the
	// command zone, the graveyard (flashback, escape), exile.
	FromZone ZoneKind

	// XValue is the value announced for {X} (CR 601.2b), needed by
	// any modifier that reasons about the spell's total mana value.
	XValue int

	// Cost is the cost as it stands at the moment this modifier is
	// consulted: after the alternative-cost swap and the commander
	// tax, after every increase for a reduction, after everything
	// for a CostFloor. A modifier that needs the spell's PRINTED
	// mana value should read Card.ManaCost instead — this field
	// moves as the pass proceeds, by design.
	Cost ParsedCost
}

// CostModifier is one "spells cost {N} more / less to cast" static
// ability contributed by a permanent on the battlefield.
//
// Declared as a struct of hooks rather than an interface for the
// same reason StaticAbility, ReplacementEffect and TriggeredAbility
// are: a card file writes a literal, not a type.
type CostModifier struct {
	// Kind places the modifier in the CR 601.2f order.
	Kind CostModifierKind

	// Label is the clause as printed ("Noncreature spells cost {1}
	// more to cast"), for the event log and for debugging a cast
	// that came out at an unexpected price.
	Label string

	// AppliesTo decides whether this modifier touches this cast. Nil
	// means "every spell", which is Sphere of Resistance exactly.
	//
	// Evaluated once per cast per modifier, under the write lock.
	// Read-only.
	AppliesTo func(q CostQuery) bool

	// Amount is the size of the modification: generic mana added
	// (CostIncrease), generic mana removed (CostReduction), or the
	// minimum total mana value (CostFloor).
	//
	// A function rather than an int because three of the six cards
	// in the sprint scale with the board: Animar counts its +1/+1
	// counters, Damping Sphere counts the spells already cast this
	// turn. A fixed modifier writes `Fixed(1)`.
	//
	// Returning a negative number refuses the cast — see the file
	// header. Nil is the same as returning zero, which is a no-op
	// rather than an error: a modifier that declares no amount
	// simply does not apply.
	Amount func(q CostQuery) int
}

// CatalogCostModifiers is the catalog hook the effects package wires
// at init, mirroring CatalogStaticAbilities. Nil, or a nil return,
// means the card modifies nobody's costs — which is nearly every
// card.
var CatalogCostModifiers func(oracleID string) []CostModifier

// CostModifiersFor returns the cost modifiers a card contributes, or
// nil.
func CostModifiersFor(oracleID string) []CostModifier {
	if CatalogCostModifiers == nil || oracleID == "" {
		return nil
	}
	return CatalogCostModifiers(oracleID)
}

// boundCostModifier is a declared modifier paired with the
// battlefield permanent that contributes it. The source is copied
// rather than referenced: the modifier pass runs inside the cast
// path, and a pointer into g.Battlefield.Cards would go stale the
// moment a modifier's own predicate caused a board change (it
// shouldn't, but "shouldn't" is not a memory-safety argument).
type boundCostModifier struct {
	modifier CostModifier
	source   Card
}

// activeCostModifiersLocked collects every cost modifier currently
// in play — one entry per (battlefield permanent, declared
// modifier) pair.
//
// Battlefield only (CR 113.6). A cost modifier that works from
// another zone exists (Trinisphere does not; the sprint's six
// increasers and three reducers are all permanents) but none of them
// is in scope, and widening the scan later is a one-line change here
// rather than a change at every call site.
//
// Caller must hold g.mu.
func (g *Game) activeCostModifiersLocked() []boundCostModifier {
	if g == nil || g.Battlefield == nil || CatalogCostModifiers == nil {
		return nil
	}
	var out []boundCostModifier
	for i := range g.Battlefield.Cards {
		src := g.Battlefield.Cards[i]
		// CatalogAbilityKey: a cost modifier is a static ability, so
		// a permanent under a CR 613.1f ability-removing effect
		// stops taxing and stops discounting.
		mods := CatalogCostModifiers(CatalogAbilityKey(src))
		for _, m := range mods {
			out = append(out, boundCostModifier{modifier: m, source: src})
		}
	}
	return out
}

// applyCostModifiersLocked prices `base` through every applicable
// modifier in CR 601.2f order and returns the total cost. `q` is the
// query the modifiers are judged against; its Cost field is
// overwritten as the pass proceeds, so callers may leave it zero.
//
// Errors — and so refuses the cast — when a modifier returns a
// negative amount, or when a CostFloor's minimum is negative. See
// the file header for why that is a refusal rather than a clamp.
//
// Caller must hold g.mu.
func (g *Game) applyCostModifiersLocked(base ParsedCost, q CostQuery) (ParsedCost, error) {
	mods := g.activeCostModifiersLocked()
	if len(mods) == 0 {
		return base, nil
	}
	cost := base
	// CR 601.2f step 3: every increase, in any order. Additive, so
	// "any order" really is any order.
	for _, bm := range mods {
		if bm.modifier.Kind != CostIncrease {
			continue
		}
		n, ok, err := amountFor(bm, q, cost)
		if err != nil {
			return ParsedCost{}, err
		}
		if ok {
			cost.Generic += n
		}
	}
	// CR 601.2f step 4: every reduction. Also additive, but each one
	// is clamped against the generic remaining as it is spent, which
	// is what keeps two {2}-reducers from making a {1}{B} spell owe
	// negative generic.
	for _, bm := range mods {
		if bm.modifier.Kind != CostReduction {
			continue
		}
		n, ok, err := amountFor(bm, q, cost)
		if err != nil {
			return ParsedCost{}, err
		}
		if ok {
			cost = reduceGeneric(cost, n)
		}
	}
	// Trinisphere last. A second floor on top of the first is a
	// no-op by construction — the cost already clears the higher
	// minimum — so two Trinispheres behave like one, as printed.
	for _, bm := range mods {
		if bm.modifier.Kind != CostFloor {
			continue
		}
		n, ok, err := amountFor(bm, q, cost)
		if err != nil {
			return ParsedCost{}, err
		}
		if ok {
			cost = raiseToMinimum(cost, n, q.XValue)
		}
	}
	return cost, nil
}

// ApplyCostModifiers is the read-locked public surface: price `base`
// through the board's cost modifiers without casting anything. Used
// by the auto-tap preview endpoint so the plan the client is shown
// is a plan for the price the cast will actually charge.
//
// Callers must NOT hold g.mu.
func (g *Game) ApplyCostModifiers(base ParsedCost, q CostQuery) (ParsedCost, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	q.Game = g
	return g.applyCostModifiersLocked(base, q)
}

// ApplyCostModifiersForEffect is the *ForEffect twin for callers
// already under g.mu — the S31 legal-move enumerator, which runs
// inside ReadSnapshot and has to price the casts it offers at the
// board's real prices or it will advertise moves the engine then
// rejects for insufficient mana.
//
// Read-only, same contract as AutoTapForCostForEffect.
func (g *Game) ApplyCostModifiersForEffect(base ParsedCost, q CostQuery) (ParsedCost, error) {
	q.Game = g
	return g.applyCostModifiersLocked(base, q)
}

// amountFor evaluates one bound modifier against the cast. Returns
// (amount, applies, error). `applies` is false when the predicate
// declined or the amount came out zero — a zero-amount modifier is
// not an error, it just does nothing (Animar with no counters,
// Damping Sphere on the turn's first spell).
func amountFor(bm boundCostModifier, q CostQuery, cost ParsedCost) (int, bool, error) {
	q.Source = bm.source
	q.Cost = cost
	if bm.modifier.AppliesTo != nil && !bm.modifier.AppliesTo(q) {
		return 0, false, nil
	}
	if bm.modifier.Amount == nil {
		return 0, false, nil
	}
	n := bm.modifier.Amount(q)
	if n < 0 {
		return 0, false, fmt.Errorf("%w: %q on %s returned %d",
			ErrCostModifier, bm.modifier.Label, bm.source.Name, n)
	}
	return n, n > 0, nil
}

// reduceGeneric spends `n` generic mana out of a cost, stopping at
// zero. Coloured requirements — including hybrid and Phyrexian ones,
// which are coloured requirements with an escape hatch, not generic
// mana — are untouched.
//
// The one rule this function exists to keep: a reduction that
// overshoots is not carried forward and does not start eating {B}.
func reduceGeneric(cost ParsedCost, n int) ParsedCost {
	if n <= 0 {
		return cost
	}
	out := cost
	if n >= out.Generic {
		out.Generic = 0
		return out
	}
	out.Generic -= n
	return out
}

// raiseToMinimum implements a cost-setting effect: if the cost's
// total mana value is below `min`, add generic mana until it is
// exactly `min`. Already-expensive spells are untouched.
//
// Trinisphere's reminder text spells out the shape — "a spell that
// would cost {1}{B} to cast costs {2}{B} to cast instead" — and the
// part worth pinning is that the coloured half survives: the
// shortfall is paid in generic, so the {B} is still a {B}.
//
// X counts toward the total at the announced value (CR 202.3b), so
// a {X}{U} spell announced with X=2 already costs three and
// Trinisphere adds nothing.
func raiseToMinimum(cost ParsedCost, min, xValue int) ParsedCost {
	if min <= 0 {
		return cost
	}
	total := cost.ManaValueWithX(xValue)
	if total >= min {
		return cost
	}
	out := cost
	out.Generic += min - total
	return out
}

// ManaValue is the cost's converted mana cost with {X} counted as
// zero (CR 202.3b — X is zero everywhere except on the stack).
func (c ParsedCost) ManaValue() int {
	return c.Generic + len(c.Required)
}

// ManaValueWithX is ManaValue with {X} counted at the announced
// value, which is what it is worth while the spell is on the stack
// and while its total cost is being determined.
func (c ParsedCost) ManaValueWithX(x int) int {
	return c.ManaValue() + c.XSlots*x
}

// ManaValue is the card's printed mana value, or zero when the cost
// can't be read. Lands parse to the zero cost and so are zero, which
// is right (CR 202.3a).
//
// Deliberately printed-cost-only: cost modifiers, alternative costs
// and the commander tax all change what a spell COSTS and none of
// them change its mana value (CR 202.3c).
func (c Card) ManaValue() int {
	cost, err := ParseCost(c.ManaCost)
	if err != nil {
		return 0
	}
	return cost.ManaValue()
}
