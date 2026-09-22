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
//	   claimed at announce (CR 601.2b, already settled by the time
//	   this file runs);
//	2. additional costs (paid, not priced — see additional_cost.go);
//	3. cost INCREASES, in any order among themselves;
//	4. cost REDUCTIONS, in any order among themselves;
//	5. cost-SETTING effects last (Trinisphere).
//
// # A spell's own cost modifier (ADR 0048 addendum, #746)
//
// Some spells change what they THEMSELVES cost: "this spell costs {1}
// less to cast for each creature on the battlefield" (Blasphemous
// Act), affinity (CR 702.41a), Fireball's "{1} more for each target
// beyond the first". Those live in their own catalog slot,
// CardDef.SelfCostModifiers, read only for the card being priced and
// never from the battlefield (SelfCostModifiersFor). They join the
// same three passes below, so every increase — from the board and
// from the spell — still lands before any reduction, and a self
// reduction floors at generic exactly like a board one.
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

	// Source is the permanent contributing the modifier — or, for a
	// self modifier, the spell being cast, with Controller set to the
	// caster (CR 601.2a). Its
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

	// Targets are the targets announced for the spell (CR 601.2c),
	// which come before the total cost is determined (CR 601.2f).
	// Filled from CastSpellParams.Targets by the cast, and per target
	// set by the legal-move enumerator.
	//
	// SHOWN ONLY TO A MODIFIER THAT SETS ReadsTargets (ADR 0048
	// addendum §13). Every other modifier sees nil here, in the
	// engine and in the enumerator alike, so a card that reads
	// targets without declaring it gets the same wrong answer in both
	// places — which its own test catches — rather than an engine
	// that charges one price and an enumerator that advertises
	// another (#544).
	Targets []TargetRef

	// Cost is the cost as it stands at the moment this modifier is
	// consulted: after the alternative-cost swap and the commander
	// tax, after every increase for a reduction, after everything
	// for a CostFloor. A modifier that needs the spell's PRINTED
	// mana value should read Card.ManaCost instead — this field
	// moves as the pass proceeds, by design.
	Cost ParsedCost

	// Ability names the ACTIVATED ABILITY being priced when this
	// query is about an activation (CR 602.2f) rather than a cast,
	// and is nil for every cast (#1184).
	//
	// It is the whole of what a cost modifier could not previously
	// see. CostQuery already carried the SOURCE of an ability
	// (Card / Controller are the permanent and its controller), so
	// "abilities of other permanents you control" was expressible;
	// what no predicate could reach was the ability itself — which
	// one of the permanent's abilities this is, and whether it
	// prints the exhaust keyword. Boom Scholar's "Exhaust abilities
	// of other permanents you control cost {2} less to activate" is
	// exactly that pair.
	//
	// Nil-ness is also the SELECTOR: a modifier declares which kind
	// of announcement it prices with CostModifier.Activations, and
	// activeCostModifiersLocked shows a cast's modifiers only casts
	// and an activation's modifiers only activations. Without that,
	// Sphere of Resistance — whose AppliesTo is nil, meaning "every
	// spell" — would have started taxing every activated ability in
	// the game the day this field appeared.
	Ability *AbilityCostSubject
}

// AbilityCostSubject is the ability an activation-pricing CostQuery
// is about. A value on the query rather than the ActivatedAbilityShape
// itself, so a predicate cannot reach the ability's Effect closure or
// its target spec and start asking questions pricing has no business
// asking.
type AbilityCostSubject struct {
	// Label is the ability's printed label — the same string the
	// activation record is keyed by (activation_tally.go).
	Label string

	// Exhaust is the exhaust keyword bit (#1181): "Activate each
	// exhaust ability only once".
	Exhaust bool

	// Mana says this is a CR 605 mana ability rather than a CR 602
	// activation. Always false today — the mana path prices its cost
	// without this pass at all — and present so that a predicate
	// written now says which kind it means rather than being silently
	// widened the day the other path joins. See ADR 0020's exhaust
	// addendum, note of 2026-09-22.
	Mana bool
}

// CostModifier is one "spells cost {N} more / less to cast" static
// ability contributed by a permanent on the battlefield — or, in a
// card's SelfCostModifiers slot, one "this spell costs {N} more /
// less to cast" ability of the spell being cast.
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

	// ReadsTargets declares that AppliesTo or Amount looks at
	// q.Targets — Fireball's "{1} more for each target beyond the
	// first", Price of Fame's "{2} less if it targets a legendary
	// creature". Without it the modifier is handed nil targets.
	//
	// It is also how the enumerator knows it must price a cast per
	// target set instead of once up front (ADR 0048 addendum §14).
	// Set by the target-reading constructors; a card file never
	// writes it by hand.
	ReadsTargets bool

	// Unit is what one point of Amount adds, for a CostIncrease whose
	// increase is not plain generic mana: strive's "{1}{W} more for
	// each target beyond the first" (Call the Coppercoats) has a Unit
	// of {1}{W}, so an Amount of 2 adds {2}{W}{W}. Nil — every other
	// modifier — means one generic mana per point.
	//
	// Increases only (ADR 0048 addendum, open question 3). A
	// reduction still spends generic mana and nothing else (§3), and
	// a Unit on any other kind, or a Unit carrying {X}, a hybrid or a
	// Phyrexian symbol, refuses the cast rather than guessing what it
	// meant (§16). effects.Register refuses the same shapes at boot;
	// both read UnitProblem.
	Unit *ParsedCost

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this modifier applies only while its source permanent has the
	// designation named. Fortune Teller's Talent's "Spells you cast
	// from anywhere other than your hand cost {2} less" is
	// ClassLevel(3). The zero value is "no gate".
	//
	// Evaluated in CostModifiersForCard and nowhere else, so a
	// gated-off modifier never reaches the CR 601.2f pass — which
	// means the engine and the legal-move enumerator price the cast
	// identically, because both read the same accessor.
	//
	// Meaningless in SelfCostModifiers: that slot is read while the
	// card is a spell being cast, where it has no designations. See
	// designations.go and ADR 0071.
	ActiveWhen Designation

	// Activations declares that this modifier prices ACTIVATED
	// ABILITY costs — "Exhaust abilities of other permanents you
	// control cost {2} less to activate" (Boom Scholar, #1184) —
	// rather than casts.
	//
	// It is a partition, not a widening: a modifier prices
	// activations or casts and never both, because every printed
	// clause in either family says which it means. "Spells cost {1}
	// more to cast" and "abilities cost {2} less to activate" are
	// different sentences, and a card that printed both would print
	// two clauses and declare two modifiers.
	//
	// Read in activeCostModifiersLocked, against CostQuery.Ability's
	// nil-ness, which is what keeps Sphere of Resistance's "every
	// spell" from meaning "every announcement". Nothing else in the
	// CR 601.2f pass changes: the order, the generic floor and the
	// negative-amount refusal are the same rules, applied to the
	// ability's mana component.
	Activations bool
}

// UnitProblem says why a modifier's Unit cannot be applied, or ""
// when it can: no Unit, or a Unit on an increase made only of generic
// mana and single-colour symbols. ADR 0048 addendum §16: no printed
// increase adds {X}, a hybrid or a Phyrexian symbol, and §3 keeps
// every other kind generic-only, so each of those shapes is a card
// file mistake. One function, so the boot guard and the cast-time
// refusal cannot drift apart.
func (m CostModifier) UnitProblem() string {
	u := m.Unit
	if u == nil {
		return ""
	}
	if m.Kind != CostIncrease {
		return "a mana unit on something other than an increase"
	}
	if u.XSlots > 0 {
		return "a mana unit carrying {X}"
	}
	if u.HasPhyrexian {
		return "a mana unit carrying a Phyrexian symbol"
	}
	for _, r := range u.Required {
		if r.Phyrexian {
			return "a mana unit carrying a Phyrexian symbol"
		}
		if len(r.Options) != 1 || r.HasNumericAlt {
			return "a mana unit carrying a hybrid symbol"
		}
	}
	return ""
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

// SelfCostModifiersFor returns the cost modifiers a card applies to
// ITSELF while it is being cast — "this spell costs {1} less to cast
// for each creature on the battlefield", affinity (CR 702.41a) — or
// nil. Read off the card's CardDef under CatalogKey, so an MDFC or
// split card answers for the face the cast stamped and no other
// (ADR 0048 addendum §12).
//
// Never read from the battlefield: a modifier in this slot changes
// what its own card costs and nobody else's, which is why it is a
// separate slot from CostModifiersFor rather than a flag on it.
func SelfCostModifiersFor(card Card) []CostModifier {
	if d := catalogDef(CatalogKey(card)); d != nil {
		return d.SelfCostModifiers
	}
	return nil
}

// TargetPricedCostClauses returns the printed clauses of the card's
// own cost modifiers that price by target — Fireball's "This spell
// costs {1} more to cast for each target beyond the first" — for the
// catalog key `key`, or nil. The client's X picker opens before
// targeting (ADR 0021 §3) and prices at one target, so it shows these
// under its readout rather than a number it cannot know yet (ADR 0048
// addendum, open question 2, option a).
func TargetPricedCostClauses(key string) []string {
	d := catalogDef(key)
	if d == nil {
		return nil
	}
	var out []string
	for _, m := range d.SelfCostModifiers {
		if m.ReadsTargets && m.Label != "" {
			out = append(out, m.Label)
		}
	}
	return out
}

// boundCostModifier is a declared modifier paired with the card that
// contributes it: a battlefield permanent, or the spell being cast
// for its own self modifiers. The source is copied rather than
// referenced: the modifier pass runs inside the cast path, and a
// pointer into g.Battlefield.Cards would go stale the moment a
// modifier's own predicate caused a board change (it shouldn't, but
// "shouldn't" is not a memory-safety argument).
type boundCostModifier struct {
	modifier CostModifier
	source   Card
}

// activeCostModifiersLocked collects every cost modifier that bears
// on this cast — one entry per (battlefield permanent, declared
// modifier) pair, then the spell's own self modifiers.
//
// The battlefield half is CR 113.6: a permanent's static works from
// the battlefield, and a Sphere of Resistance in a hand taxes
// nobody. That is why the scan was never simply widened to every
// zone (ADR 0048 addendum, alternatives considered).
//
// The self half is CR 113.6d: an ability that modifies what its own
// object costs to cast functions on the stack, so it applies whatever
// zone the spell is cast from — hand, command zone, graveyard, exile.
// It is bound with the card being cast as its source and the caster
// as that source's controller (CR 601.2a), so a YourSpell() predicate
// on a self modifier reads true rather than comparing against a hand
// card's zero controller. Self modifiers come after the battlefield
// ones within each pass, for determinism; §2's passes are additive,
// so the order changes no result.
//
// Caller must hold g.mu.
func (g *Game) activeCostModifiersLocked(q CostQuery) []boundCostModifier {
	if g == nil {
		return nil
	}
	var out []boundCostModifier
	if g.Battlefield != nil && CatalogCostModifiers != nil {
		for i := range g.Battlefield.Cards {
			src := g.Battlefield.Cards[i]
			// CostModifiersForCard: a cost modifier is a static
			// ability, so a permanent under a CR 613.1f
			// ability-removing effect stops taxing and stops
			// discounting (CatalogAbilityKey), and one whose
			// designation gate is unsatisfied — Fortune Teller's
			// Talent below level 3 — is not there at all (ADR 0071).
			mods := CostModifiersForCard(src)
			for _, m := range mods {
				// #1184: casts are priced by the cast modifiers and
				// activations by the activation ones, and never the
				// other way. See CostModifier.Activations.
				if m.Activations != (q.Ability != nil) {
					continue
				}
				out = append(out, boundCostModifier{modifier: m, source: src})
			}
		}
	}
	// #1184: an ability's cost is not its source's cost. The self
	// slot is "THIS SPELL costs {N} less to cast" (CR 113.6d) and is
	// about the card as a spell on the stack; a permanent's activated
	// ability is a different announcement by a different object, so
	// an affinity-carrying artifact does not discount its own tap
	// ability.
	if q.Ability != nil {
		return out
	}
	if self := SelfCostModifiersFor(q.Card); len(self) > 0 {
		src := q.Card
		src.Controller = q.Controller
		for _, m := range self {
			out = append(out, boundCostModifier{modifier: m, source: src})
		}
	}
	return out
}

// CastPriceReadsTargetsForEffect reports whether pricing a cast of
// `card` right now could depend on the targets chosen for it: the
// card's own self modifiers, or any modifier on the battlefield,
// declares ReadsTargets. The legal-move enumerator asks it to decide
// whether one up-front price is enough or each target set has to be
// priced on its own (ADR 0048 addendum §14).
//
// Deliberately ignores AppliesTo, so it errs toward true: a false
// answer is a promise that targets cannot move the price.
//
// Caller must hold g.mu (read or write).
func (g *Game) CastPriceReadsTargetsForEffect(card Card) bool {
	for _, m := range SelfCostModifiersFor(card) {
		if m.ReadsTargets {
			return true
		}
	}
	if g == nil || g.Battlefield == nil || CatalogCostModifiers == nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		for _, m := range CatalogCostModifiers(CatalogAbilityKey(g.Battlefield.Cards[i])) {
			if m.ReadsTargets {
				return true
			}
		}
	}
	return false
}

// RealTargetCount is the number of announced targets that name
// something, ignoring the TargetSelf and TargetNone placeholders,
// counted exactly as the cast's own target validator counts them.
// "For each target beyond the first" is this number minus one.
func RealTargetCount(targets []TargetRef) int {
	return countRealTargets(targets)
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
	mods := g.activeCostModifiersLocked(q)
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
			cost = increaseBy(cost, n, bm.modifier.Unit)
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

// abilityCostQueryLocked builds the CostQuery for one activation of
// `ab` by `activator` from `source`. One builder, so the payment path
// and the legal-move enumerator cannot ask the board a differently
// shaped question about the same click (#544, #1184).
//
// Caller must hold g.mu.
func (g *Game) abilityCostQueryLocked(activator uuid.UUID, source Card, zone ZoneKind, ab ActivatedAbilityShape) CostQuery {
	return CostQuery{
		Game: g,
		// The permanent (or, since #660, the hand card) whose ability
		// this is, and the player activating it — the "you" of the
		// activation, as against the modifier source's controller,
		// which is the "you" of the modifier's own clause.
		Card:       source,
		Controller: activator,
		// Where the SOURCE is — the battlefield for nearly every
		// ability, the hand for a cycling one (#660, CR 113.6). Passed
		// in rather than looked up: both callers already know it,
		// because both had to check it to get here.
		FromZone: zone,
		Ability: &AbilityCostSubject{
			Label:   ab.Label,
			Exhaust: ab.Exhaust,
		},
	}
}

// AbilityManaCostForEffect is what one activation of `ab` will
// actually charge in mana: the printed component, priced through
// every activation-scoped cost modifier on the battlefield in
// CR 601.2f order (#1184).
//
// ONE function and three readers, which is the same discipline
// Game.AbilityExhausted holds the exhaust gate to:
//
//  1. ActivateCatalogAbility, which pays this and no other number;
//  2. internal/legal, which decides whether the seat can afford the
//     move it is about to offer — a bot offered an activation the
//     engine then refuses for want of mana is #544 exactly;
//  3. the ability view (#1190), rendered through ParsedCost.String()
//     into ActivatedAbilityView.ChargedManaCost, so the row the
//     player clicks and the number the engine takes are the same
//     string.
//
// Errors exactly where the cast path errors: an unparseable printed
// cost, or a modifier that returns a nonsensical amount. It does not
// clamp and does not fall back to the printed cost — #289's lesson,
// one path over.
//
// Caller must hold g.mu (read or write).
func (g *Game) AbilityManaCostForEffect(activator uuid.UUID, source Card, zone ZoneKind, ab ActivatedAbilityShape) (ParsedCost, error) {
	base, err := ParseCost(ab.Cost.Mana)
	if err != nil {
		return ParsedCost{}, err
	}
	if ab.Cost.Mana == "" {
		// Nothing to discount and nothing to tax: a cost with no mana
		// component is not made of mana, so the pass has no subject.
		// The early return is also what keeps the battlefield walk
		// off the {T}-only abilities that are most of the catalog.
		return base, nil
	}
	return g.applyCostModifiersLocked(base, g.abilityCostQueryLocked(activator, source, zone, ab))
}

// manaAbilityCostQueryLocked builds the CostQuery for one activation of
// the CR 605 mana ability `ab` on `source` by `activator` — the mana
// twin of abilityCostQueryLocked (#1191), for the same reason: the
// manual activation path, the legal-move enumerator and the
// auto-tapper's planner must ask the board the same shaped question
// about the same click.
//
// Mana: true is the one difference from a CR 602 query — #1193's
// AbilityCostSubject.Mana, always false until this file set it,
// existing for exactly this so a predicate written before today (or
// tomorrow) says which kind of activation it means rather than being
// silently widened.
//
// A mana ability is always a battlefield permanent's — ManaAbilitiesForCard
// never reads a hand card — so FromZone is always the battlefield,
// unlike abilityCostQueryLocked's, which a caller passes in for #660's
// sake.
//
// Caller must hold g.mu.
func (g *Game) manaAbilityCostQueryLocked(activator uuid.UUID, source Card, ab ManaAbilityShape) CostQuery {
	return CostQuery{
		Game:       g,
		Card:       source,
		Controller: activator,
		FromZone:   ZoneBattlefield,
		Ability: &AbilityCostSubject{
			Label:   ab.Label,
			Exhaust: ab.Exhaust,
			Mana:    true,
		},
	}
}

// ManaAbilityManaCostForEffect is AbilityManaCostForEffect's twin for a
// CR 605 mana ability (#1191): the printed mana component of `ab`'s
// own activation cost — the Signet cycle's "{1}, {T}", Loot, the
// Pathfinder's exhaust "{G}, {T}" — priced through every
// activation-scoped cost modifier on the battlefield in CR 601.2f
// order.
//
// CR 605.1a makes a mana ability an activated ability, so Boom
// Scholar's "Exhaust abilities of other permanents you control cost
// {2} less to activate" reaches Loot's mana ability exactly as it
// reaches a CR 602 one — through the SAME activeCostModifiersLocked
// partition (CostModifier.Activations against CostQuery.Ability's
// nil-ness), unchanged by this function. Nothing about the order, the
// generic floor or the negative-amount refusal is mana-specific;
// what was missing was a query to hand the pass, which this builds.
//
// Three readers, mirroring AbilityManaCostForEffect's:
// ActivateManaAbility, which pays this and no other number; the
// legal-move enumerator's manaMoves, so a bot is never offered a mana
// activation the engine then refuses for want of mana (#544); and the
// auto-tapper (autoTapAbilityFor / materializePlanLocked), which must
// know a discounted mana component from a printed one before it can
// decide whether a source is plannable.
//
// Caller must hold g.mu (read or write).
func (g *Game) ManaAbilityManaCostForEffect(activator uuid.UUID, source Card, ab ManaAbilityShape) (ParsedCost, error) {
	base, err := ParseCost(ab.ManaCost)
	if err != nil {
		return ParsedCost{}, err
	}
	if ab.ManaCost == "" {
		// Nothing to discount and nothing to tax, exactly as the CR
		// 602 early return: a tap-only mana ability is not made of
		// mana, so the pass has no subject and the battlefield walk
		// never runs for the ordinary land or Sol Ring.
		return base, nil
	}
	return g.applyCostModifiersLocked(base, g.manaAbilityCostQueryLocked(activator, source, ab))
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
	// §13: targets are shown only to a modifier that says it reads
	// them, so the engine and the enumerator can never disagree about
	// a modifier that forgot to say so.
	if !bm.modifier.ReadsTargets {
		q.Targets = nil
	}
	// A Unit is an increase's shape (open question 3). On any other
	// kind, or carrying an {X} no announcement could size, or a hybrid
	// or Phyrexian symbol whose payment no printed increase defines
	// (§16), there is no honest price: refuse, as for a negative
	// amount.
	if why := bm.modifier.UnitProblem(); why != "" {
		return 0, false, fmt.Errorf("%w: %q on %s declares %s",
			ErrCostModifier, bm.modifier.Label, bm.source.Name, why)
	}
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

// increaseBy adds `n` units of an increase to a cost: `n` generic
// mana when unit is nil, and otherwise `n` copies of the unit's
// generic and coloured requirements — strive's {1}{W} twice is
// {2}{W}{W}. The coloured symbols join Required as ordinary
// requirements, which the payment solver already pays.
func increaseBy(cost ParsedCost, n int, unit *ParsedCost) ParsedCost {
	if n <= 0 {
		return cost
	}
	if unit == nil {
		cost.Generic += n
		return cost
	}
	out := cost
	out.Generic += n * unit.Generic
	if len(unit.Required) > 0 {
		req := make([]ColorRequirement, 0, len(cost.Required)+n*len(unit.Required))
		req = append(req, cost.Required...)
		for i := 0; i < n; i++ {
			req = append(req, unit.Required...)
		}
		out.Required = req
	}
	// A Phyrexian or hybrid unit never gets here: amountFor refuses it
	// through UnitProblem before the increase pass runs.
	out.HasSnow = cost.HasSnow || unit.HasSnow
	return out
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
// X counts toward the total at the announced value (CR 202.3e), so
// a {X}{U} spell announced with X=2 already costs three and
// Trinisphere adds nothing.
//
// The floor is measured against the TOTAL COST (CR 601.2f), not the
// mana value, and the two differ on a monocoloured hybrid symbol: the
// player announces which half of a {2/W} they will pay (CR 601.2b),
// and the engine's payment path always pays the coloured half, so a
// {2/W} costs one mana here while it is worth two to the mana value
// (CR 202.3f). See totalManaWithX.
func raiseToMinimum(cost ParsedCost, min, xValue int) ParsedCost {
	if min <= 0 {
		return cost
	}
	total := cost.totalManaWithX(xValue)
	if total >= min {
		return cost
	}
	out := cost
	out.Generic += min - total
	return out
}

// ManaValue is the cost's mana value (CR 202.3) with {X} counted as
// zero (CR 202.3e — X is zero everywhere except on the stack). It is
// the one place the engine turns a parsed cost into a mana value;
// every card-side read goes through it, via Card.ManaValue,
// Card.ParsedManaValue or Card.ManaValueWithX.
//
// Symbol by symbol:
//
//   - {N} contributes N; {X} contributes 0 (CR 202.3e).
//   - {W} {U} {B} {R} {G} {C} and snow {S} contribute 1 each.
//   - A two-colour hybrid {W/U}, and a colourless hybrid {C/W},
//     contribute 1: the largest component is one mana (CR 202.3f).
//   - A monocoloured hybrid {2/W} contributes 2, its larger component
//     (CR 202.3f — "{2/B}{2/B}{2/B} is 6").
//   - Phyrexian {W/P} contributes 1 (CR 202.3g).
//
// This is the mana VALUE, not what the cast charges — for that see
// totalManaWithX, which counts a {2/W} as the one coloured mana the
// payment path actually takes.
func (c ParsedCost) ManaValue() int {
	mv := c.Generic
	for _, r := range c.Required {
		mv += r.ManaValue()
	}
	return mv
}

// ManaValueWithX is ManaValue with {X} counted at the announced
// value, which is what it is worth while the spell is on the stack
// (CR 202.3e).
func (c ParsedCost) ManaValueWithX(x int) int {
	return c.ManaValue() + c.XSlots*x
}

// totalManaWithX is the amount of mana the engine charges for the
// cost with {X} at the announced value: generic, plus X, plus one per
// Required slot. It deliberately counts a monocoloured hybrid {2/W}
// as ONE — the payment path pays every hybrid with its coloured half
// (see ColorRequirement), which is the nonhybrid equivalent the
// engine announces on the caster's behalf (CR 601.2b). A cost-setting
// effect (Trinisphere) measures this, not the mana value.
func (c ParsedCost) totalManaWithX(x int) int {
	return c.Generic + c.XSlots*x + len(c.Required)
}

// ManaValue is the card's printed mana value, or zero when the cost
// can't be read. Lands parse to the zero cost and so are zero, which
// is right (CR 202.3a).
//
// Deliberately printed-cost-only: cost modifiers, alternative costs
// and the commander tax all change what a spell COSTS and none of
// them change its mana value (CR 202.3).
func (c Card) ManaValue() int {
	mv, _ := c.ParsedManaValue()
	return mv
}

// ParsedManaValue is ManaValue that also reports whether the printed
// cost could be read. A predicate that must not match a card whose
// cost the engine cannot price ("target creature with mana value 3 or
// less" against a joined split-card cost) reads ok rather than
// treating the unreadable cost as zero, which would make every such
// card pass a ceiling it might not meet.
func (c Card) ParsedManaValue() (mv int, ok bool) {
	cost, err := ParseCost(manaCostForValue(c))
	if err != nil {
		return 0, false
	}
	return cost.ManaValue(), true
}

// manaCostForValue is the cost a MANA VALUE is computed from, which is
// the card's own printed cost for everything but one case.
//
// CR 712.8e: "While a nonmodal double-faced permanent has its back
// face up, it has only the characteristics of its back face. However,
// its mana value is calculated using the mana cost of its FRONT face."
//
// A `transform` back face prints no cost at all, and ParseCost("")
// succeeds as the zero cost — so without this, every permanent that
// transformed would have mana value 0. Reflection of Kiki-Jiki is mana
// value 3 off Fable of the Mirror-Breaker's {2}{R}, and "destroy target
// permanent with mana value 3 or less" has to agree.
//
// `transform` ONLY. CR 712.8f gives a modal DFC each face its own
// characteristics, cost included, so Sea Gate, Reborn really is mana
// value 0 and reading its front face here would make a land cost seven.
// Added in S46 (ADR 0079 decision 8, #343).
func manaCostForValue(c Card) string {
	if c.Layout == LayoutTransform && c.ActiveFace != 0 && len(c.Faces) > 0 {
		return c.Faces[0].ManaCost
	}
	return c.ManaCost
}

// ManaValueWithX is the card's mana value as a spell on the stack,
// with {X} counted at x, the value chosen for it (CR 202.3e). Zero
// when the cost can't be read, like ManaValue.
func (c Card) ManaValueWithX(x int) int {
	cost, err := ParseCost(manaCostForValue(c))
	if err != nil {
		return 0
	}
	return cost.ManaValueWithX(x)
}

// ManaValueForEffect is the card's mana value where it is right now.
// It is the read to use for any card that might be a spell on the
// stack: a counterspell's "target spell with mana value N", a "whenever
// you cast a spell with mana value N or greater" trigger, the limit a
// cascade records. CR 202.3e: on the stack, {X} counts as the value
// chosen for it (a copy has the X of the spell it copies, CR 707.10).
// Everywhere else, {X} is zero, so for a card in any other zone this
// gives the same answer as Card.ParsedManaValue.
//
// ok is false when the printed cost can't be read, like
// Card.ParsedManaValue. A predicate should reject such a card rather
// than treat it as mana value zero.
//
// "On the stack" means the card is in the stack zone AND has a spell
// entry in StackMeta. The entry holds the announced X. A card that
// has left the stack keeps its InstanceID, but its entry is dropped
// when it leaves, so X no longer counts. Safe on a nil game, which
// reads X as zero.
//
// Caller must hold g.mu, like every other *ForEffect read.
func (g *Game) ManaValueForEffect(c Card) (mv int, ok bool) {
	cost, err := ParseCost(manaCostForValue(c))
	if err != nil {
		return 0, false
	}
	return cost.ManaValueWithX(g.announcedXOnStackLocked(c.InstanceID)), true
}

// announcedXOnStackLocked is the X chosen for the spell `id` while it
// is on the stack, and zero for anything that isn't a spell on the
// stack (CR 202.3e).
func (g *Game) announcedXOnStackLocked(id uuid.UUID) int {
	if g == nil || g.Stack == nil || g.StackMeta == nil {
		return 0
	}
	item := g.StackMeta[id]
	if item == nil || item.XValue <= 0 || !g.Stack.Contains(id) {
		return 0
	}
	return item.XValue
}
