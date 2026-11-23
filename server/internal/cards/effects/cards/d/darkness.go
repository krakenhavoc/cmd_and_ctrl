package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darkness — Instant {B}:
//
//	"Prevent all combat damage that would be dealt this turn."
//
// Fog for one black mana (#3213, the Kuja deck). The shield is the
// same turn-scoped replacement Fog registers, so it covers damage to
// creatures and players alike and clears at cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "932708ae-9d5c-4561-aa8b-0d2222d37fdc",
		Name:         "Darkness",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PreventAllCombatDamageThisTurn{
				Label: "Darkness: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
