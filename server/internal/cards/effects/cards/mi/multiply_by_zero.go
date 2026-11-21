package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Multiply by Zero — Instant {1}{B}:
//
//	"Target creature has base power and toughness 0/0 until end of
//	 turn."
//
// A layer 7b "base P/T becomes 0/0", so counters and later +N/+N
// effects still apply on top of it (a creature with two +1/+1 counters
// ends as a 2/2, one with none dies to the 0-toughness state-based
// action).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d940611f-f84d-43fa-82d6-eecb4fb54164",
		Name:         "Multiply by Zero",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, item.Targets[0].ID, nil,
				"Multiply by Zero — base power and toughness 0/0 until end of turn",
				game.SetBasePTMods(0, 0)...)
		},
	})
}
