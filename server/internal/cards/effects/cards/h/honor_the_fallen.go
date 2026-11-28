package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Honor the Fallen — Instant {1}{W} (Mercadian Masques, S82 OLD, #5619):
//
//	"Exile all creature cards from all graveyards. You gain 1 life for each
//	 card exiled this way."
//
// Every seat's graveyard is read as the spell resolves and the creature cards
// are exiled as one batch. The life is one per card that really reached exile, so
// a commander its owner returned to the command zone instead gives none, and it
// follows once any such choice has been made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db557d5e-eecd-4891-90fa-c5bcb8532015",
		Name:         "Honor the Fallen",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ids := effects.AllGraveyardsCardIDs(ctx, game.Card.IsCreature)
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.ExileCardsKeyedForEffect(ids, ctx.ItemTail(honorTheFallenExiled))
		},
	})
}

// honorTheFallenExiled is "you gain 1 life for each card exiled this way". An
// on-disk key: never renamed, never reused.
var honorTheFallenExiled = effects.CardTail("honor-the-fallen/gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.GainLife{Player: ctx.Controller(), Amount: len(o.IDs)}.Apply(ctx)
})
