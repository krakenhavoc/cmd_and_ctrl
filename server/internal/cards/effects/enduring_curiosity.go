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
			On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
				return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
			}, "Enduring Curiosity — return it to the battlefield; it's an enchantment, not a creature",
				enduringCuriosityReturnAsEnchantment),
		},
	})
}

// enduringCuriosityReturnAsEnchantment is the dies trigger's Effect: a
// package-level func, not a closure, so it captures nothing — undo
// restores a cloned game and this has to resolve against whichever one
// it is handed. item.SourceCardID already carries the instance to
// move (NewTriggeredItem stamps it from the ability's source), and the
// returned permanent keeps that SAME instance ID: CR 400.7's "new
// object" is tracked here by the (InstanceID, EnteredBattlefieldAt)
// pair the duration model already reads (see
// sameObjectOnBattlefieldLocked in game/duration.go), not by minting a
// fresh one on every zone change.
func enduringCuriosityReturnAsEnchantment(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id := item.SourceCardID
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
		// CR 608.2b: something moved the card out of the graveyard in
		// response — exiled it, put it back on top of a library, a
		// second effect reanimated it first — and the trigger still
		// resolves; it just has nothing left to bring back.
		return nil
	}
	if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
		return err
	}
	affected := ctx.Game.PinnedObjectsLocked(id)
	if len(affected) == 0 {
		// It didn't actually take the battlefield — a replacement or a
		// state-based action swept it away the instant it entered —
		// so there is nothing left to keep from being a creature.
		return nil
	}
	// A data record (ADR 0041 phase 3, #1497): the effect lasts as long
	// as the permanent does, and as a closure it kept the table off the
	// restore path for all of it. Registered straight onto the game,
	// not through ScopedEffectFor: the permanent IS this trigger's
	// source, returned as a new object, which ScopedEffectFor's "this"
	// guard (#1432) would rightly refuse to call "this".
	ctx.Game.RegisterScopedEffectForEffect(ctx.Source(), affected,
		[]game.Mod{game.RemoveTypesMod("Creature")},
		ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
		"Enduring Curiosity — it's an enchantment (it's not a creature)")
	return nil
}
