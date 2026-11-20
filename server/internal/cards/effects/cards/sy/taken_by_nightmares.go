package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Taken by Nightmares — Instant {2}{B}{B} (Wilds of Eldraine, #4471):
//
//	"Exile target creature. If you control an enchantment, scry 2."
//
// Angelic Edict's exile, then Roadside Reliquary's "if you control an
// enchantment", asked of the board as the spell resolves, after the
// exile (CR 608.2c): an enchantment you control then scries, whatever
// happened to the creature. The scry does not depend on the exile, so a
// creature a replacement kept from exile still lets you scry. A target
// that left in response leaves the spell with no legal target, and it
// does nothing at all (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5557c9f-e4c2-41af-9fcc-371409e1dec6",
		Name:         "Taken by Nightmares",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.ExileTheFirstTargetIfLegal(ctx); err != nil {
				return err
			}
			if effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.IsEnchantment() }) == 0 {
				return nil
			}
			return effects.Scry{N: 2}.Apply(ctx)
		},
	})
}
