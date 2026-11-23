package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spark of Creativity — Sorcery {R} (Jumpstart 2022, #5948):
//
//	"Choose target creature. Exile the top card of your library. You may
//	 have Spark of Creativity deal damage to that creature equal to the
//	 exiled card's mana value. If you don't, you may play that card until
//	 end of turn."
//
// The creature is the spell's one target (a target that has become illegal
// fizzles the spell, CR 608.2b). The exile is Riverwheel Sweep's plain exile
// with a continuation, run once the card has landed in exile (an empty
// library exiles nothing and asks nothing). The "may" is asked then: a yes
// is the damage, equal to the exiled card's mana value as it is in exile
// (a land is 0, X is 0); a no is the play permission until end of turn,
// "play" so that a land can be played with the land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22504fd0-0353-4ef0-baf4-4b7be28d1a66",
		Name:         "Spark of Creativity",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			tail := ctx.ItemTail(sparkOfCreativityExiled)
			tail.Carry = []uuid.UUID{target}
			return ctx.Game.ExileTopKeyedForEffect(item.Controller, item.Controller, 1, tail)
		},
	})
}

// sparkOfCreativityExiled asks the "you may" over the exiled card. An
// on-disk key: never renamed, never reused.
var sparkOfCreativityExiled = effects.CardTail("spark-of-creativity/exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil || z.Kind != game.ZoneExile {
			continue
		}
		return effects.MayChoice{
			Question: "Spark of Creativity — deal damage to that creature equal to the exiled card's mana value? (If you don't, you may play that card this turn.)",
			YesLabel: "Deal damage",
			NoLabel:  "Play the card",
			Then:     sparkOfCreativityAnswer,
			Carry:    []uuid.UUID{o.Tail.Carry[0], id},
		}.Apply(ctx)
	}
	return nil
})

var sparkOfCreativityAnswer = effects.MayYesNo("spark-of-creativity/damage-or-play",
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) < 2 {
			return nil
		}
		c, ok := ctx.Game.LookupCardForEffect(r.Carry[1])
		if !ok {
			return nil
		}
		return effects.DealDamage{Source: ctx.Source(), Target: r.Carry[0], Amount: c.ManaValue()}.Apply(ctx)
	},
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if len(r.Carry) < 2 {
			return nil
		}
		z := ctx.Game.FindCardZoneForEffect(r.Carry[1])
		c, ok := ctx.Game.LookupCardForEffect(r.Carry[1])
		if !ok || z == nil || z.Kind != game.ZoneExile {
			return nil
		}
		perm := effects.ExileTopWithPermission{}.Permission()
		perm.Player = ctx.Controller()
		perm.Zone = game.ZoneExile
		ctx.Game.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
		return nil
	})
