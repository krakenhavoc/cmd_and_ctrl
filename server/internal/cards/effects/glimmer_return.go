package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// glimmer_return.go — the Duskmourn Enduring cycle's shared dies
// trigger:
//
//	"When <this> dies, if it was a creature, return it to the
//	 battlefield under its owner's control. It's an enchantment. (It's
//	 not a creature.)"
//
// Written once for Enduring Curiosity and shared by Enduring Tenacity,
// Enduring Vitality and Enduring Innocence.
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
// "It's an enchantment. (It's not a creature.)" is ADR 0210's entry
// overwrite (EntersAs{Types: ["Enchantment"]}, #4512): the card is an
// enchantment from the moment it lands (CR 614.1c), so it is never a
// creature on the battlefield. "Whenever a creature enters" (Soul
// Warden) does not see it, and "creatures your opponents control enter
// tapped" (Authority of the Consuls) does not apply to it (CR 614.12).
// The type is SET (CR 205.1a): every card in the cycle is an Enchantment
// Creature, so it becomes a plain Enchantment, and the creature types
// (Cat, Glimmer) go with the card type (CR 205.1a, #2041). The record is
// pinned to the returned object and swept with it, with no stated
// duration (CR 611.2a). The card keeps its own abilities.
//
// Before #4512 this registered RemoveTypesMod("Creature") AFTER the card
// had landed, so EventETB went out while it was still a creature.

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
		z := g.FindCardZoneForEffect(item.SourceCardID)
		if z == nil || z.Kind != game.ZoneGraveyard || !sameGraveyardObject(z, item) {
			// CR 608.2b: something moved the card out of the graveyard
			// in response, and the trigger has nothing left to bring
			// back. Anything that came back since is a new object the
			// trigger does not name (CR 400.7).
			return nil
		}
		// Controller zero: under its owner's control.
		return ReturnFromGraveyardTogether{
			Targets: []uuid.UUID{item.SourceCardID},
			EntersAs: &game.EntersAs{
				Types: []string{"Enchantment"},
				Label: name + " — it's an enchantment (it's not a creature)",
			},
		}.Apply(NewContext(g, item))
	}
}
