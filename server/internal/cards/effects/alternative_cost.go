package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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

// Escape is "Escape—{cost}, Exile N other cards from your graveyard.
// (You may cast this card from your graveyard for its escape cost.)"
// — CR 702.144.
//
// Flashback's sibling and its opposite in the one place that
// matters. Both are alternative costs bound to the graveyard, but
// flashback carries ExileOnLeavingStack and escape does NOT: an
// escaped permanent goes to the battlefield and its card goes to the
// graveyard the next time it dies, ready to escape again. That is
// not an oversight in the card, it is the card — escape is a
// recursion engine whose brake is the yard it eats, which is why the
// exile clause is a COST rather than a rider.
//
// The count lives on the spec (Min == Max == n), so the engine's
// count check, the client's picker and the payment all read one
// number. "Other" is enforced by the cast path rather than by the
// spec: CR 601.2a moves the spell to the stack before costs are
// paid, so the card exiling its own graveyard cannot reach itself.
//
// The card file still opens the zone — the constructor cannot,
// because CastableZones and AlternativeCosts are separate fields:
//
//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
//	AlternativeCosts: []game.AlternativeCost{Escape("{2}{B}", 4)},
func Escape(cost string, n int) game.AlternativeCost {
	clause := "Exile " + numberWord(n) + " other cards from your graveyard"
	return game.AlternativeCost{
		Key:                "escape",
		Label:              "Escape—" + cost + ", " + clause,
		ManaCost:           cost,
		FromZone:           game.ZoneGraveyard,
		ExileFromGraveyard: CardsInYourGraveyard(n, clause),
		PayLabel:           numberWord(n) + " other cards from your graveyard",
	}
}

// EscapeWithCounters is Escape plus "this creature escapes with N
// +1/+1 counters on it" (CR 702.144c) — the rider most escape
// creatures print, and the reason an escaped Voracious Typhon is a
// 7/7 rather than the 4/4 in the corner.
//
// A separate constructor rather than a variadic on Escape, because
// the counters are the half a card file forgets: a Typhon written
// with plain Escape compiles, casts, and enters as a vanilla 4/4 for
// seven mana and four cards — strictly worse than printed, which is
// the one direction the catalog is allowed to err in but a silent
// wrong answer all the same.
//
// The counters land only when the escape cost was paid. The same
// creature reanimated, blinked or hard-cast enters with none,
// because the clause hangs off the cost rather than off the card.
func EscapeWithCounters(cost string, n, counters int) game.AlternativeCost {
	ac := Escape(cost, n)
	ac.EntersWithCounterName = "+1/+1"
	ac.EntersWithCounterCount = counters
	return ac
}

// numberWord spells a small count the way an oracle line does —
// "five other cards", not "5 other cards". Falls back to digits past
// the range any printed escape cost uses.
func numberWord(n int) string {
	words := []string{"zero", "one", "two", "three", "four", "five",
		"six", "seven", "eight", "nine", "ten"}
	if n >= 0 && n < len(words) {
		return words[n]
	}
	return strconv.Itoa(n)
}

// Warp is "Warp {cost} (You may cast this card from your hand for
// its warp cost. Exile this creature at the beginning of the next
// end step, then you may cast it from exile on a later turn.)" —
// CR 702.183, and the fix for #324.
//
// Unlike flashback, warp is paid from HAND: it is a discount now in
// exchange for the real card later, which is why it needs no
// CastableZones declaration. The later cast is an ordinary cast from
// exile for the printed mana cost, riding the same
// ExilePlayPermission impulse exile and airbend already use — so
// the client's existing exile button renders it with no new code.
//
// The constructor bundles the exile clause for the same reason
// Evoke bundles its sacrifice: a card file that wrote
// `game.AlternativeCost{ManaCost: "{R}"}` by hand would ship a
// creature that costs one mana and stays on the battlefield forever,
// which is not a discount but a strictly better card.
func Warp(cost string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:       "warp",
		Label:     "Warp " + cost,
		ManaCost:  cost,
		WarpExile: true,
	}
}

// --- S28: the free-spell family ----------------------------------
//
// Four more shapes of "rather than pay this spell's mana cost", all
// on the same AlternativeCost the S22 keywords ride. What they add is
// a CONDITION ("if you control a commander") and NON-MANA components
// (pitch a card, pay life, bounce a land) — which is why they are
// constructors here rather than a parallel mechanism: the announce
// gate, the target rewrite, the wire key and the view projection all
// already exist, and none of them needed to learn a new word.

// FreeIfYouControlCommander is the Commander Legends free-spell
// cycle's clause: "If you control a commander, you may cast this
// spell without paying its mana cost." Fierce Guardianship, Deadly
// Rollick, Deflecting Swat and their three siblings.
//
// The condition is checked in the view as well as at announce, so a
// player with no commander on the battlefield is never shown the
// offer rather than being shown one the server would reject.
//
// "You control a commander" means a commander PERMANENT you control,
// which is not the same as owning one: a commander in the command
// zone does not count, and a commander you have stolen does.
func FreeIfYouControlCommander(label string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:       "free",
		Label:     label,
		ManaCost:  "",
		Condition: controlsACommander,
	}
}

// controlsACommander is the cycle's condition: the caster controls at
// least one permanent that is somebody's commander.
func controlsACommander(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCommander && c.Controller == controller {
			return true
		}
	}
	return false
}

// Pitch is the Force of Will clause: "You may pay N life and exile a
// <spec> card from your hand rather than pay this spell's mana cost."
//
// Both halves are COSTS and both are validated before either is paid,
// so a player at 1 life with no blue card gets a rejected cast rather
// than a dead player and a spell still in hand. The exile happens with
// the spell already on the stack (CR 601.2a before 601.2h), which is
// why a countered Force of Will still costs you the pitched card —
// the part a resolution-time implementation would get backwards.
func Pitch(label string, life int, from *game.TargetSpec, payLabel string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:           "pitch",
		Label:         label,
		ManaCost:      "",
		Life:          life,
		ExileFromHand: from,
		PayLabel:      payLabel,
	}
}

// EvokePitch is evoke with a card, not mana: "Evoke—Exile a white
// card from your hand" (Solitude and the rest of the Modern Horizons
// 2 incarnation cycle).
//
// Evoke() with a mana cost and this share a key, and deliberately: it
// is the same keyword, and a card offers one or the other, never
// both. What it keeps from Evoke() is the half a card file would
// forget — SacrificeOnEntry, the triggered ability that makes the
// creature enter, fire its ETB and only then die (CR 702.74b). Without
// it Solitude would be a five-mana Swords to Plowshares that stays on
// the battlefield.
func EvokePitch(from *game.TargetSpec, payLabel string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:              "evoke",
		Label:            "Evoke—Exile " + payLabel,
		ManaCost:         "",
		ExileFromHand:    from,
		PayLabel:         payLabel,
		SacrificeOnEntry: true,
	}
}

// PayLifeInstead is Snuff Out's clause: "If you control a Swamp, you
// may pay 4 life rather than pay this spell's mana cost."
//
// `condition` may be nil for an unconditional version; Snuff Out's is
// the Swamp check.
func PayLifeInstead(label string, life int, condition func(g *game.Game, controller uuid.UUID) bool) game.AlternativeCost {
	return game.AlternativeCost{
		Key:       "pay_life",
		Label:     label,
		ManaCost:  "",
		Life:      life,
		Condition: condition,
	}
}

// ReturnInstead is Daze's clause: "You may return an Island you
// control to its owner's hand rather than pay this spell's mana
// cost."
//
// The bounce is a cost, so it happens at announce with the Daze
// already on the stack — the Island is back in hand before the spell
// Daze is answering has resolved, and before its controller decides
// whether to pay the {1}.
func ReturnInstead(label string, from *game.TargetSpec, payLabel string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:          "return",
		Label:        label,
		ManaCost:     "",
		ReturnToHand: from,
		PayLabel:     payLabel,
	}
}

// ControlsA builds a "if you control a <subtype>" condition — Snuff
// Out's Swamp. Reads effective subtypes, so a land animated or
// type-changed into a Swamp counts.
func ControlsA(subtype string) func(g *game.Game, controller uuid.UUID) bool {
	return func(g *game.Game, controller uuid.UUID) bool {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == controller && hasSubtype(c, subtype) {
				return true
			}
		}
		return false
	}
}
