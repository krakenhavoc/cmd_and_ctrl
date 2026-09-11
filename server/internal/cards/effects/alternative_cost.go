package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// alternative_cost.go — S22: constructors for Spec.AlternativeCosts,
// the "you may cast this spell for its <keyword> cost rather than its
// mana cost" clause (CR 118.9).
//
// One constructor per keyword rather than a generic builder, because
// each keyword bundles a rewrite with its price and the bundling is
// the part a card file should not have to remember: overload also
// deletes the target clause, evoke also attaches a sacrifice trigger,
// cleave also swaps the clause for a wider one. A card that wrote
// `game.AlternativeCost{ManaCost: "{4}{R}"}` by hand would compile,
// cast for four, and still demand a target — a strictly worse
// Vandalblast that looks right.
//
// The Key strings are the wire contract: they ride cast_spell as
// `alternative_cost` and land on StackItem.AltCost, where a card's
// OnResolve reads them back through ctx.PaidAltCost.

// Overload is "Overload {cost} (You may cast this spell for its
// overload cost. If you do, change 'target' in its text to 'each.')"
// — CR 702.96.
//
// The cost is only half of it. Overload deletes the target clause, so
// an overloaded spell announces with no targets at all: nothing to
// pick, nothing for hexproof or protection to stop, and no fizzle if
// the board empties in response. The card's OnResolve branches on
// ctx.PaidAltCost("overload") and sweeps instead of targeting.
func Overload(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:           "overload",
		Label:         "Overload " + cost,
		ManaCost:      cost,
		ClearsTargets: true,
	}
}

// Evoke is "Evoke {cost} (You may cast this spell for its evoke cost.
// If you do, it's sacrificed when it enters.)" — CR 702.74.
//
// The sacrifice is a triggered ability, not part of the resolution,
// which is what makes evoke worth having: the creature really does
// enter, so its enters-the-battlefield trigger fires, and the
// sacrifice goes on the stack where opponents can respond to it. For
// Slithermuse — whose payoff is a LEAVES-the-battlefield trigger —
// the ordering is the entire card.
func Evoke(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:              "evoke",
		Label:            "Evoke " + cost,
		ManaCost:         cost,
		SacrificeOnEntry: true,
	}
}

// Cleave is "Cleave {cost} (You may cast this spell for its cleave
// cost. If you do, remove the words in square brackets.)" — CR
// 702.148.
//
// `targets` is the clause the spell has WITH the bracketed words
// removed, which is almost always wider than the printed one (Wash
// Away goes from "counter target spell that wasn't cast from its
// owner's hand" to "counter target spell"). Pass nil for a cleave
// whose brackets don't touch the targeting clause; the printed clause
// then stands.
func Cleave(cost string, targets *game.TargetSpec) game.AlternativeCost {
	return game.AlternativeCost{
		Key:      "cleave",
		Label:    "Cleave " + cost,
		ManaCost: cost,
		Targets:  targets,
	}
}

// Flashback is "Flashback {cost} (You may cast this card from your
// graveyard for its flashback cost. Then exile it.)" — CR 702.34.
//
// The first alternative cost that is also a cast PATH. It carries
// two things a card file must not be trusted to remember separately:
//
//   - FromZone. The offer is claimable only out of the graveyard,
//     and a card that declares it must also list ZoneGraveyard in
//     Spec.CastableZones (Register panics otherwise). Without the
//     binding the cost would be claimable from hand, which on Deep
//     Analysis or a Past-in-Flames-fuelled Past in Flames is a real
//     discount rather than a cosmetic one.
//   - ExileOnLeavingStack. This is what keeps flashback from being
//     infinite, and it is a REPLACEMENT (CR 702.34a says "any time
//     it would leave the stack"), not an exile appended to the
//     resolution — so a flashed-back spell that fizzles is exiled,
//     and one answered by Hinder is exiled rather than shuffled
//     away. A card file that wrote the cost by hand would get a
//     spell that flashes back, lands in the graveyard, and flashes
//     back again every turn forever.
//
// The card file still has to open the zone; the constructor cannot
// do it, because CastableZones and AlternativeCosts are separate
// fields on the Spec:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
func Flashback(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:                 "flashback",
		Label:               "Flashback " + cost,
		ManaCost:            cost,
		FromZone:            game.ZoneGraveyard,
		ExileOnLeavingStack: true,
	}
}
