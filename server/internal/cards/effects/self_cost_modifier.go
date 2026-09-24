package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// self_cost_modifier.go — ADR 0048 addendum (#746): the vocabulary
// for Spec.SelfCostModifiers, "THIS spell costs {N} more / less to
// cast" (CR 601.2f, 113.6d).
//
// The ordinary constructors in cost_modifier.go (CostsLess,
// CostsLessEach, CostsMore, CostsMoreEach) work in the self slot
// unchanged: the SLOT is what makes a modifier apply to its own
// spell. This file adds what only a spell's own modifier needs:
//
//   - AffinityFor — CR 702.41a, "costs {1} less for each [text] you
//     control". One entry per instance, so two instances both apply
//     (CR 702.41b).
//   - CostsMorePerTargetBeyondFirst — Fireball's "{1} more" and
//     strive's "{1}{W} more for each target beyond the first".
//   - CostsLessIfItTargets — Price of Fame's "{2} less if it targets a
//     legendary creature".
//   - the counting amounts the printed clauses use.
//
// Two rules every amount here keeps:
//
//   - It EXCLUDES the card being cast (q.Card). The engine prices
//     before it moves the card, so at pricing time the card still
//     sits in its source zone; under CR 601.2a it is already on the
//     stack. "For each card in your graveyard" read during a
//     flashback cast must not count the card being cast.
//   - It never returns a negative number. A negative amount refuses
//     the cast (ADR 0048 §4), and "the total power of creatures you
//     control" can be negative on paper, where Ghalta's ruling reads
//     it as no reduction at all.
//
// The two target-reading constructors set CostModifier.ReadsTargets
// themselves. A card file never sets it by hand: a modifier that
// reads q.Targets without the flag is handed nil, in the engine and
// in the legal-move enumerator alike (§13).

// AffinityFor is "Affinity for [text]" (CR 702.41a): this spell costs
// {1} less to cast for each permanent you control that matches.
// Thought Monitor is AffinityFor("Affinity for artifacts", Artifact()).
//
// The label is the printed keyword line; the reminder text is the
// same for every instance and adds nothing to the log.
func AffinityFor(label string, match CardPredicate) game.CostModifier {
	return CostsLessEach(PermanentsYouControl(match), label)
}

// CostsMorePerTargetBeyondFirst is "This spell costs {per} more to
// cast for each target beyond the first" — Fireball ("{1}") and the
// strive ability word ("{1}{W}", Call the Coppercoats; CR 207.2c).
//
// Zero or one target adds nothing, so an unpriced query with no
// targets reads the one-target price and never trips the negative-
// amount refusal. `per` may carry coloured symbols: the increase pass
// adds them as ordinary coloured requirements (ADR 0048 addendum,
// open question 3). It may not carry {X}. Panics at init on a string
// the cost parser rejects, which is where a typo belongs.
func CostsMorePerTargetBeyondFirst(per string, label string) game.CostModifier {
	unit, err := game.ParseCost(per)
	if err != nil || unit.XSlots > 0 || unit.ManaValue() == 0 {
		panic(fmt.Sprintf("effects.CostsMorePerTargetBeyondFirst: %q is not a fixed mana amount (%v)", per, err))
	}
	return game.CostModifier{
		Kind:  game.CostIncrease,
		Label: label,
		Amount: func(q game.CostQuery) int {
			if n := game.RealTargetCount(q.Targets); n > 1 {
				return n - 1
			}
			return 0
		},
		ReadsTargets: true,
		Unit:         &unit,
	}
}

// CostsLessIfItTargets is "This spell costs {n} less to cast if it
// targets [a card that matches]" — Price of Fame's "a legendary
// creature". `match` is an ordinary target predicate, judged for the
// caster against the announced card targets as they stand when the
// cost is determined (CR 601.2f, after 601.2c).
func CostsLessIfItTargets(n int, label string, match CardPredicate, when ...CostPredicate) game.CostModifier {
	preds := append(append([]CostPredicate(nil), when...), ItTargets(match))
	m := CostsLess(n, label, preds...)
	m.ReadsTargets = true
	return m
}

// ItTargets passes when one of the spell's announced targets is a
// card that matches. It reads q.Targets, so it is only meaningful on
// a modifier that sets ReadsTargets — use it through
// CostsLessIfItTargets, which does.
func ItTargets(match CardPredicate) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, t := range q.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			c, ok := q.Game.LookupCardForEffect(t.ID)
			if ok && match(q.Game, q.Controller, c) {
				return true
			}
		}
		return false
	}
}

// CostsLessForTheCardItTargets is "This ability costs {1} less to
// activate for each [thing about] the creature it targets" — Dragonfire
// Blade's colours, Warrior's Blades' +1/+1 counters (#1296). `per`
// reads the targeted card, as it stands when the cost is determined
// (CR 602.2b: 601.2c's targets come before 601.2f's total), and says
// how many generic mana come off.
//
// It reads the FIRST card target, because every printed clause of
// this family says "the creature it targets" on an ability with one
// target. No card target — a query priced before one is chosen, or a
// target that has left — reads zero, which for a reduction is the
// printed cost: the direction #259 prefers when there is nothing to
// count.
//
// Sets ReadsTargets, so the engine hands it the announced targets and
// the enumerator and the view price per target. Declare it on
// ActivatedAbility.CostModifiers; nothing stops it pricing a spell
// from Spec.SelfCostModifiers, but no printed spell says this.
func CostsLessForTheCardItTargets(label string, per func(c game.Card) int) game.CostModifier {
	return game.CostModifier{
		Kind:  game.CostReduction,
		Label: label,
		Amount: func(q game.CostQuery) int {
			if q.Game == nil {
				return 0
			}
			for _, t := range q.Targets {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := q.Game.LookupCardForEffect(t.ID)
				if !ok {
					return 0
				}
				if n := per(c); n > 0 {
					return n
				}
				return 0
			}
			return 0
		},
		ReadsTargets: true,
	}
}

// ColorsOf counts a card's colours as they are right now (layer 5 on
// the battlefield, CR 105.2): a Vivi Ornitier is two, an artifact
// creature none. Dragonfire Blade's "for each color of the creature it
// targets".
func ColorsOf(c game.Card) int {
	return len(c.EffectiveColors())
}

// CountersOf returns a counter-counting reader for
// CostsLessForTheCardItTargets — Warrior's Blades' "for each +1/+1
// counter on the creature it targets".
func CountersOf(kind string) func(c game.Card) int {
	return func(c game.Card) int {
		return c.Counters[kind]
	}
}

// --- counting amounts -------------------------------------------------

// PermanentsYouControl counts the permanents the CASTER controls that
// match — affinity's "for each artifact you control", Hamza's "for
// each creature you control with a +1/+1 counter on it". Reads the
// live battlefield, so an animated land or a Mycosynth Lattice'd
// creature counts by its effective types.
func PermanentsYouControl(match CardPredicate) func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		return countBattlefield(q, func(c game.Card) bool {
			return c.Controller == q.Controller && (match == nil || match(q.Game, q.Controller, c))
		})
	}
}

// PermanentsOnBattlefield counts every permanent that matches, whoever
// controls it — Blasphemous Act and Vanquish the Horde's "for each
// creature on the battlefield".
func PermanentsOnBattlefield(match CardPredicate) func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		return countBattlefield(q, func(c game.Card) bool {
			return match == nil || match(q.Game, q.Controller, c)
		})
	}
}

// TotalPowerOfCreaturesYouControl is Ghalta, Primal Hunger's "X is
// the total power of creatures you control": current power, so
// counters and anthems count. A total of zero or less reduces
// nothing (Ghalta's ruling), which also keeps the amount from ever
// being negative.
func TotalPowerOfCreaturesYouControl() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		total := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.InstanceID == q.Card.InstanceID || c.Controller != q.Controller || !c.IsCreature() {
				continue
			}
			total += c.CurrentPower()
		}
		if total < 0 {
			return 0
		}
		return total
	}
}

// OpponentsOfCaster counts the caster's opponents still in the game —
// undaunted's "{1} less for each opponent" (CR 702.125a).
func OpponentsOfCaster() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		n := 0
		for _, p := range q.Game.Seats {
			if p != nil && !p.Eliminated && p.ID != q.Controller {
				n++
			}
		}
		return n
	}
}

// LifeBelowStartingTotal is Shadow of Mortality's "if your life total
// is less than your starting life total, … X is the difference", and
// zero otherwise.
func LifeBelowStartingTotal() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		p := q.Game.PlayerByIDForEffect(q.Controller)
		if p == nil || p.Life >= game.StartingLife {
			return 0
		}
		return game.StartingLife - p.Life
	}
}

// CountCardsInYourGraveyard counts the cards in the caster's graveyard that
// match, never counting the card being cast — Rumbleweed's "for each
// land card in your graveyard", The Magic Mirror's "for each instant
// and sorcery card".
func CountCardsInYourGraveyard(match CardPredicate) func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil {
			return 0
		}
		p := q.Game.PlayerByIDForEffect(q.Controller)
		if p == nil || p.Graveyard == nil {
			return 0
		}
		n := 0
		for _, c := range p.Graveyard.Cards {
			if c.InstanceID == q.Card.InstanceID {
				continue
			}
			if match == nil || match(q.Game, q.Controller, c) {
				n++
			}
		}
		return n
	}
}

// countBattlefield is the shared walk under the two permanent counts.
// Reads the battlefield in place — the cast path holds the lock and
// this runs once per modifier per cast — and skips the card being
// cast, which is never a permanent while it is being cast.
func countBattlefield(q game.CostQuery, keep func(game.Card) bool) int {
	if q.Game == nil || q.Game.Battlefield == nil {
		return 0
	}
	n := 0
	for _, c := range q.Game.Battlefield.Cards {
		if c.InstanceID != q.Card.InstanceID && keep(c) {
			n++
		}
	}
	return n
}

// --- conditions -----------------------------------------------------

// YouControlA passes when the caster controls a permanent that
// matches — Winged Words' "if you control a creature with flying",
// Wizard's Retort's "if you control a Wizard".
func YouControlA(match CardPredicate) CostPredicate {
	return func(q game.CostQuery) bool {
		return PermanentsYouControl(match)(q) > 0
	}
}

// PermanentsOnBattlefieldAtLeast passes when at least n permanents on
// the battlefield match — Hour of Revelation's "if there are ten or
// more nonland permanents on the battlefield".
func PermanentsOnBattlefieldAtLeast(n int, match CardPredicate) CostPredicate {
	return func(q game.CostQuery) bool {
		return PermanentsOnBattlefield(match)(q) >= n
	}
}

// AnOpponentHasCardsInGraveyardAtLeast passes when some opponent of the
// caster has at least n cards in their graveyard — Into the Story.
func AnOpponentHasCardsInGraveyardAtLeast(n int) CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, p := range q.Game.Seats {
			if p == nil || p.Eliminated || p.ID == q.Controller || p.Graveyard == nil {
				continue
			}
			if p.Graveyard.Size() >= n {
				return true
			}
		}
		return false
	}
}

// YouGainedLifeThisTurn passes when the caster has gained life this
// turn — Mortality Spear.
func YouGainedLifeThisTurn() CostPredicate {
	return func(q game.CostQuery) bool {
		return q.Game != nil && q.Game.TurnTallyFor(q.Controller).LifeGained > 0
	}
}

// AnOpponentControlsNoBasicLands passes when at least one of the
// caster's opponents controls no basic land — Hagra Mauling.
func AnOpponentControlsNoBasicLands() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil || q.Game.Battlefield == nil {
			return false
		}
		for _, p := range q.Game.Seats {
			if p == nil || p.Eliminated || p.ID == q.Controller {
				continue
			}
			hasBasic := false
			for _, c := range q.Game.Battlefield.Cards {
				if c.Controller == p.ID && IsBasicLand(c) {
					hasBasic = true
					break
				}
			}
			if !hasBasic {
				return true
			}
		}
		return false
	}
}

// ArtifactCreatureSpell passes on a spell that is both an artifact and
// a creature — Mycosynth Golem's "artifact creature spells you cast".
func ArtifactCreatureSpell() CostPredicate {
	return func(q game.CostQuery) bool { return q.Card.IsArtifact() && q.Card.IsCreature() }
}

// --- card predicates the counts above use ---------------------------

// CreatureWithCounter matches a creature with at least one counter of
// the named kind on it — Hamza's "creature you control with a +1/+1
// counter on it" (pair it with PermanentsYouControl for the "you
// control").
func CreatureWithCounter(kind string) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.Counters[kind] > 0
	}
}

// AttackingCreature matches a creature that is currently attacking —
// Ancient Stone Idol's "for each attacking creature". Combat
// declarations are cleared as the end of combat step ENDS (CR 511.3),
// so outside combat nothing matches.
func AttackingCreature() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.IsCreature() && c.AttackingTarget != uuid.Nil
	}
}

// TotalManaValueOfHistoricPermanentsYouControl is Excalibur, Sword of
// Eden's "X is the total mana value of historic permanents you
// control" — CR 700.6's historic (an artifact, a legendary, or a
// Saga), summed rather than counted.
//
// Reads the PRINTED mana cost of each permanent, which is what mana
// value is (CR 202.3): a token has none and contributes nothing, and
// a land contributes nothing unless it is somehow an artifact or
// legendary with a cost, which no land is.
func TotalManaValueOfHistoricPermanentsYouControl() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		total := 0
		for _, c := range q.Game.Battlefield.Cards {
			if c.InstanceID == q.Card.InstanceID || c.Controller != q.Controller {
				continue
			}
			if b09IsHistoric(c) {
				total += c.ManaValue()
			}
		}
		if total < 0 {
			return 0
		}
		return total
	}
}
