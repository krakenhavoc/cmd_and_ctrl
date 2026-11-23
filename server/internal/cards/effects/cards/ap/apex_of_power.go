package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apex of Power — Sorcery {7}{R}{R}{R}:
//
//	"Exile the top seven cards of your library. Until end of turn, you
//	 may cast spells from among them.
//	 If this spell was cast from your hand, add ten mana of any one
//	 color."
//
// The grant is cast-only (a land among them is not playable, as the
// text says "cast spells") and lapses at end of turn. The mana is read
// off the zone the spell was cast from (CR 601.2a), after the exile, as
// printed; it is ten of ONE color the controller picks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d06eedb4-bd5a-4864-9505-a570a745d165",
		Name:         "Apex of Power",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Impulse: 7, Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := (effects.ExileTopWithPermission{From: ctx.Controller(), GrantTo: ctx.Controller(), N: 7, CastOnly: true}).Apply(ctx); err != nil {
				return err
			}
			if item.CastFromZone == game.ZoneHand {
				return effects.AddMana{Produced: effects.OneColorOfAmount(10)}.Apply(ctx)
			}
			return nil
		},
	})
}
