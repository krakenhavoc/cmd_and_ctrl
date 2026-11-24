package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Possibility Storm — Enchantment {3}{R}{R} (Secret Lair Drop, #5462):
//
//	"Whenever a player casts a spell from their hand, that player exiles
//	 it, then exiles cards from the top of their library until they exile a
//	 card that shares a card type with it. That player may cast that card
//	 without paying its mana cost. Then they put all cards exiled with this
//	 enchantment on the bottom of their library in a random order."
//
// One trigger over every player's casts from their hand (the cast event's
// OldZone, CR 601.2a). Its steps run in the printed order. The spell is
// exiled from the stack without being countered (ExileTargetSpell; a spell
// that already left in response is not exiled and the walk still happens, from
// the card's types as they were), and once that has settled the caster's
// library is walked to the first card that shares a card type with it
// (CR 205.2, SharedCardTypes: the same card types the spell had as it was
// cast). The walk is cascade's body (Game.ExileUntilCastFreeForEffect):
// the caster may cast the hit free during this resolution (CR 608.2g), a land
// can't be cast so it is passed over, and every card exiled this way that was
// not cast, the spell included, goes to the bottom of the caster's library in
// one random order. A walk that finds nothing bottoms all of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8586169f-cb8c-4424-ba64-a7c1cc00680e",
		Name:         "Possibility Storm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, possibilityStormCastFromHand,
				"Possibility Storm — that player exiles it, then exiles cards from the top of their library until they exile a card that shares a card type with it, and may cast that card without paying its mana cost",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if item.Trigger == nil {
						return nil
					}
					return effects.ExileTargetSpell{
						StackID:  item.Trigger.Event.CardID,
						ThenTail: ctx.ItemTail(possibilityStormWalk),
					}.Apply(ctx)
				}),
		},
	})
}

// possibilityStormCastFromHand is "a player casts a spell from their hand".
func possibilityStormCastFromHand(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.OldZone == game.ZoneHand && ev.Actor != uuid.Nil
}

// possibilityStormWalk is everything after the spell's exile: the walk, the
// free cast and the bottoming. An on-disk key: never renamed, never reused.
var possibilityStormWalk = effects.CardTail("possibility-storm/walk", func(ctx *effects.Context, o game.TailOutcome) error {
	item := ctx.Item
	if item == nil || item.Trigger == nil {
		return nil
	}
	spellID, player := item.Trigger.Event.CardID, item.Trigger.Event.Actor
	spell, ok := ctx.Game.LookupCardForEffect(spellID)
	if !ok {
		return nil
	}
	var also []uuid.UUID
	for _, id := range o.IDs {
		if id == spellID {
			also = append(also, id)
		}
	}
	return ctx.Game.ExileUntilCastFreeAlsoBottomingForEffect(player, ctx.Source(), func(c game.Card) bool {
		return len(game.SharedCardTypes([]game.Card{spell, c})) > 0
	}, also)
})
