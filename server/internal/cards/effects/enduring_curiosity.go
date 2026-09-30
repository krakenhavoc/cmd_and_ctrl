package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Curiosity — Enchantment Creature — Cat Glimmer, {2}{U}{U},
// 4/3 (#321, #1107):
//
//	"Flash
//	 Whenever a creature you control deals combat damage to a player,
//	 draw a card.
//	 When Enduring Curiosity dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// Three ordinary-looking clauses, none of which turned out to be:
//
//   - Flash is PrintedKeywords, same as any other card.
//   - The draw is "A creature", not "one or more creatures" — Old
//     Gnawbone's shape, not Keeper of Fables' — so no OncePerBatch:
//     three attackers connecting is three draws.
//   - The dies half is the reason this file exists. "If it was a
//     creature" is CR 603.4's intervening-if, checked once against
//     the CR 603.10 last-known-information the harvester hands
//     AppliesTo as `lki`, not re-checked at resolution — it is a fact
//     about the object's PAST characteristics, not a live board
//     condition, so there is nothing left to re-check by the time the
//     trigger would resolve.
//
// The dies half is shared with Enduring Tenacity and lives in
// glimmer_return.go (WhenThisDiesReturnItAsAnEnchantment).
//
// "It's an enchantment. (It's not a creature.)" is a CR 613.3 type
// change with no stated duration (CR 611.2a): it is about THIS
// RETURN, so it follows the permanent that came back and ends with
// it. That is an indefinite duration PINNED to the returned object
// (ADR 0041 phase 3's data record, #1497) — not a "for as long as"
// duration, which CR 702.26f would end the moment the permanent phases
// out; the pin survives a phase cycle (CR 702.26d) and is swept only
// when the object is gone. Pinned to the returned permanent's own
// instance and entry stamp, so it cannot leak onto some other Enduring
// Curiosity at the table or onto a fresh hard-cast of this one.
//
// The effect removes ONLY "Creature" from Types and leaves Subtypes
// untouched — the printed text doesn't say "loses all other types"
// (contrast Song of the Dryads' SetsBasicLandType, which does), so Cat
// and Glimmer stay on the type line same as any noncreature
// permanent's leftover creature types. If it dies again later, it
// dies as an enchantment: `lki.Types` no longer has "Creature", the
// intervening-if is false, and the ability simply doesn't trigger a
// second time — which is also the printed card's own answer.
//
// Closes #321: the card was never in the catalog (the bug's own
// triage comment confirms this — "no file in
// server/internal/cards/effects/, zero grep hits"), so there was no
// trigger to misfire. Both halves it flagged as blocked (the
// combat-damage trigger, and layer 4 not yet reading Effective()) have
// since shipped; this is the card being written now that both are
// real.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9d2460c3-8eeb-4f35-b6f6-748c478664c7",
		Name:            "Enduring Curiosity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			}, "Enduring Curiosity — draw a card", Do(DrawCards{N: 1})),
			WhenThisDiesReturnItAsAnEnchantment("Enduring Curiosity"),
		},
	})
}
