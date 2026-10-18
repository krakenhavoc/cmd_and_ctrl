package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Endless Evil — Enchantment — Aura {2}{U}:
//
//	"Enchant creature you control
//	 At the beginning of your upkeep, create a token that's a copy of
//	 enchanted creature, except the token is 1/1.
//	 When enchanted creature dies, if that creature was a Horror,
//	 return this card to its owner's hand."
//
// The copy is CreateTokenCopy with an exception that sets the token's
// printed power and toughness to 1 (CR 707.9b — the exception is part
// of the copiable values, so a later pump or counter still stacks on
// top of it). It is made under the Aura's controller, which is also the
// enchanted creature's controller by the "you control" clause; if the
// creature is gone or the Aura is not attached, there is nothing to
// copy and nothing happens.
//
// The dies trigger is harvested while the Aura is still attached (the
// dies event is emitted before the state-based action that puts the
// fallen-off Aura into the graveyard, CR 704.5m), and its intervening
// "if that creature was a Horror" (CR 603.4) is judged on the creature
// AS IT LAST EXISTED on the battlefield (CR 603.10a, leftAsSubtype), so
// a Horror by a layer effect that ended with it still counts and a
// creature that was only a Horror in the graveyard does not. The
// condition is fixed by last-known information, so nothing needs to be
// re-read at resolution.
//
// "This card" is the Aura in its owner's graveyard, where the state-
// based action put it; it returns only if it is still that object
// (CR 400.7), so an Aura exiled or returned in response stays where it
// went (sourceCardInGraveyardAfterLeaving, as Avatar Destiny does).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05157e8e-4c8c-455e-a5a1-de8271951c51",
		Name:         "Endless Evil",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Endless Evil — create a 1/1 token copy of enchanted creature", endlessEvilCopy),
			On(game.EventLTB, endlessEvilHorrorDied,
				"Endless Evil — return this card to its owner's hand", endlessEvilReturn),
		},
	})
}

// endlessEvilHorrorDied is the dies trigger's condition: the enchanted
// creature died and was a Horror as it last existed.
func endlessEvilHorrorDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !equippedCreatureDied(ev, source) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	// "Horror" is a word of the Aura's text (CR 612.2, #3984).
	return ok && leftAsSubtype(ev, c, source.TextSubtype("Horror"))
}

// endlessEvilCopy is the upkeep body: a 1/1 token copy of the enchanted
// creature for the Aura's controller.
func endlessEvilCopy(g *game.Game, item *game.StackItem) error {
	host := attachedHostFor(g, item.SourceCardID)
	if host == nil {
		return nil
	}
	return CreateTokenCopy{
		Controller: item.Controller,
		Copy:       host.InstanceID,
		N:          1,
		Except: func(t *game.Card) {
			t.Power = 1
			t.Toughness = 1
			t.VariableToughness = false
		},
	}.Apply(NewContext(g, item))
}

// endlessEvilReturn puts the Aura back into its owner's hand.
func endlessEvilReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if !sourceCardInGraveyardAfterLeaving(ctx) {
		return nil
	}
	return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(ctx)
}
