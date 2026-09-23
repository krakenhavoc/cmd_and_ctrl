package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tap_cost.go — S22: constructors for Spec.TapCost, the "tap
// permanents you control to help pay for this" cost component that
// convoke and waterbend share (game/tap_cost.go has the engine side).
//
// One constructor per keyword rather than a generic builder, for the
// same reason alternative_cost.go has three: each keyword bundles a
// pool of legal permanents and a colour rule with its name, and the
// bundling is the part a card file should not have to remember. A
// card that hand-wrote `game.TapPermanentsCost{ColorClause: true}`
// for waterbend would compile and quietly ship a card that pays
// coloured mana it should not.
//
// Untapped is baked into both specs rather than left to the
// validator, because the spec is also what the client's picker is
// built from: a tapped creature in the list would be an option the
// server then refuses.

// Untapped passes for permanents that are not currently tapped — the
// only ones that can pay a tap cost. Lives here rather than in
// targets.go because no TARGET clause in the catalog asks for it;
// "target untapped creature" would, and this is ready if one lands.
func Untapped() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return !c.Tapped }
}

// Convoke is "Convoke (Your creatures can help cast this spell. Each
// creature you tap while casting this spell pays for {1} or one mana
// of that creature's color.)" — CR 702.51.
//
// The colour clause is the half that makes convoke more than a
// discount: a tapped white creature can pay the {W} in {2}{W}{W},
// not just a generic {1}. The engine honours it by matching tapped
// creatures against the cost's coloured symbols before letting the
// leftovers pay generic.
//
// Tapping for convoke is not the {T} symbol (CR 702.51b), so a
// creature that entered this turn may still be tapped for it.
func Convoke() *game.TapPermanentsCost {
	return &game.TapPermanentsCost{
		Key:         "convoke",
		Label:       "Convoke",
		ColorClause: true,
		Spec: TargetCreature("an untapped creature you control",
			YouControl(), Untapped()),
	}
}

// Waterbend is "Waterbend {cost}. (While paying a waterbend cost,
// you can tap your artifacts and creatures to help. Each one pays
// for {1}.)"
//
// Convoke with a different name and no colour clause — each tapped
// permanent pays exactly {1} — plus one structural difference that
// matters: a waterbend cost is a cost OF ITS OWN, layered on top of
// whatever the card already charges, and the tapping pays that cost
// rather than the card's. Tapping three creatures for Waterbender's
// Restoration's waterbend {X} does not make its {U}{U} any cheaper.
// `cost` is that extra demand in brace notation — "{X}" for a
// waterbend whose size the caster announces, "{4}" for a fixed one.
//
// The pool is wider than convoke's in exchange: artifacts as well as
// creatures, so a Sol Ring can pay for {1} here and can't there.
func Waterbend(cost string) *game.TapPermanentsCost {
	return &game.TapPermanentsCost{
		Key:   "waterbend",
		Label: "Waterbend " + cost,
		Extra: cost,
		Spec: TargetPermanent("an untapped artifact or creature you control",
			Or(Artifact(), Creature()), YouControl(), Untapped()),
	}
}

// WaterbendCost is "Waterbend {cost}:" as an ACTIVATED ability's cost
// (#1310, CR 701.67a) — Aang, Swift Savior's "Waterbend {8}: Transform
// Aang", Katara, Water Tribe's Hope's "Waterbend {X}", Avatar Kuruk's
// "Exhaust — Waterbend {20}".
//
// Two halves, and a card file writes neither by hand:
//
//   - the MANA is the whole cost ("Pay [cost]"), so it goes in the
//     ability's mana component, where the cost-modifier pass, the X
//     prompt and every affordability check already read it;
//   - the Waterbend clause is Waterbend(cost) above, unchanged, whose
//     Extra names the part of that mana the taps may cover.
//
// On a spell the waterbend is an additional cost and Extra is ADDED
// to what the card costs; here it is the cost, so it is already
// there. Composed with a mana component by Plus, the two mana strings
// are summed rather than the later one winning, because CR 701.67b
// makes the waterbend cost part of the total — and the taps still
// pay only the waterbend's own generic, never the rest:
//
//	Plus(ManaCost("{1}{U}"), WaterbendCost("{2}"))  // "{1}{U}, Waterbend {2}"
//
// X composes with MinX exactly as a mana {X} does:
// Plus(WaterbendCost("{X}"), MinX(1)) is Katara's "X can't be 0".
func WaterbendCost(cost string) game.AbilityCost {
	return game.AbilityCost{Mana: cost, Waterbend: Waterbend(cost)}
}
