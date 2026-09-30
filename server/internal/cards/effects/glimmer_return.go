package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// glimmer_return.go — the Duskmourn Enduring cycle's shared dies
// trigger:
//
//	"When <this> dies, if it was a creature, return it to the
//	 battlefield under its owner's control. It's an enchantment. (It's
//	 not a creature.)"
//
// Written once for Enduring Curiosity and shared with Enduring
// Tenacity; the other Enduring cards print the same sentence.
//
// "If it was a creature" is CR 603.4's intervening-if, checked once
// against the CR 603.10 last-known information the harvester hands
// AppliesTo as `lki`. It is a fact about the object's PAST
// characteristics, so there is nothing to re-check at resolution. A
// Glimmer that dies a second time dies as an enchantment, the
// condition is false, and it stays in the graveyard — the printed
// card's own answer, and what keeps the cycle from being an
// unkillable loop.
//
// "It's an enchantment. (It's not a creature.)" is a CR 613.3 type
// change with no stated duration (CR 611.2a): it follows the
// permanent that came back and ends with it. That is an indefinite
// duration PINNED to the returned object (ADR 0041 phase 3's data
// record, #1497) — not a "for as long as" duration, which CR 702.26f
// would end the moment the permanent phases out; the pin survives a
// phase cycle (CR 702.26d) and is swept only when the object is gone.
// Only "Creature" leaves the type list; the creature subtypes stay,
// as on any noncreature permanent with leftover creature types.

// WhenThisDiesReturnItAsAnEnchantment is the whole printed trigger for
// the card named `name`.
func WhenThisDiesReturnItAsAnEnchantment(name string) game.TriggeredAbility {
	return On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, _ *game.Game) bool {
		return cardDied(ev, source) && keywordSliceContains(lki.Types, "Creature")
	}, name+" — return it to the battlefield; it's an enchantment, not a creature",
		returnItAsAnEnchantment(name))
}

// returnItAsAnEnchantment is the trigger's Effect. It captures only
// the card's name, so it resolves against whichever game it is handed
// (undo restores a clone). item.SourceCardID carries the instance to
// move, and the returned permanent keeps that instance ID: CR 400.7's
// "new object" is tracked by the (InstanceID, EnteredBattlefieldAt)
// pair the duration model reads (sameObjectOnBattlefieldLocked in
// game/duration.go).
func returnItAsAnEnchantment(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := item.SourceCardID
		if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
			// CR 608.2b: something moved the card out of the graveyard
			// in response, and the trigger has nothing left to bring
			// back.
			return nil
		}
		if err := (ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}).Apply(ctx); err != nil {
			return err
		}
		affected := ctx.Game.PinnedObjectsLocked(id)
		if len(affected) == 0 {
			// It didn't take the battlefield — a replacement or a
			// state-based action swept it away as it entered — so there
			// is nothing left to keep from being a creature.
			return nil
		}
		// Registered straight onto the game, not through
		// ScopedEffectFor: the permanent IS this trigger's source,
		// returned as a new object, which ScopedEffectFor's "this"
		// guard (#1432) would rightly refuse to call "this".
		ctx.Game.RegisterScopedEffectForEffect(ctx.Source(), affected,
			[]game.Mod{game.RemoveTypesMod("Creature")},
			ctx.Game.PinnedTo(game.IndefiniteDuration(), id),
			name+" — it's an enchantment (it's not a creature)")
		return nil
	}
}
