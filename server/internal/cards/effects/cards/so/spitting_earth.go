package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spitting Earth — Sorcery {1}{R} (Jumpstart, #5510):
//
//	"Spitting Earth deals damage to target creature equal to the number of
//	 Mountains you control."
//
// The Mountains are counted as the spell resolves (CR 608.2h), read through
// the spell's own text for the land type (CR 612.2). A target gone by then
// makes the spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c3bf4e1-d91e-4dd3-a800-e40971222c71",
		Name:         "Spitting Earth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Damage: 1}),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := effects.B08LandsWithSubtypeControlled(ctx.Game, ctx.Controller(), ctx.TextSubtype("Mountain"))
			return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: n}.Apply(ctx)
		},
	})
}
