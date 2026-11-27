package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Genesis — Enchantment {4}{B}{B} (Prophecy, S82 OLD batch 675, #5617):
//
//	"At the beginning of each player's upkeep, that player mills a card. Then
//	 they create X 1/1 black Minion creature tokens, where X is the milled
//	 card's mana value."
//
// An every-upkeep trigger (Tombstone Stairwell's) whose "that player" is the
// upkeep's player, read off the event. The mill is the keyed one-card mill
// (MillToZone), and "the milled card" is the card that really reached the
// graveyard: a card a replacement sent elsewhere, or an empty library, mills
// nothing and so makes nothing. X is its mana value as it sits in the
// graveyard ({X} in its cost counts as 0 there, CR 202.3e). The tokens are
// that player's, however many (a doubler applies).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c3ce3473-e3bc-4b7c-ae57-957e6c40dc7a",
		Name:         "Infernal Genesis",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.AtEachUpkeep("Infernal Genesis — that player mills a card, then creates X 1/1 black Minion tokens, where X is the milled card's mana value",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					ctx := effects.NewContext(g, item)
					then := ctx.ItemTail(infernalGenesisMilled)
					then.Carry = []uuid.UUID{player}
					return effects.MillToZone{Player: player, N: 1, ThenTail: then}.Apply(ctx)
				}), game.Purpose{Tokens: 1, CreatureTokens: 1}),
		},
	})
}

// infernalGenesisMilled is "Then they create X tokens": the milled card's
// mana value, for the player whose upkeep it is. An on-disk key: never
// renamed, never reused.
var infernalGenesisMilled = effects.CardTail("then/s82-old-b675-infernal-genesis-milled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 || len(o.IDs) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	x, _ := ctx.Game.ManaValueForEffect(c)
	if x <= 0 {
		return nil
	}
	return effects.CreateToken{Controller: o.Tail.Carry[0], Template: effects.TokenCard("1/1 black Minion"), N: x}.Apply(ctx)
})
