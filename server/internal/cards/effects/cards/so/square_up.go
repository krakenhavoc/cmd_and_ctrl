package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Square Up — Instant {1}{G/U} (Strixhaven: School of Mages, S82 batch 047,
// #5502):
//
//	"Target creature has base power and toughness 4/4 until end of turn."
//
// Multiply by Zero's shape with 4/4: a layer 7b "base power and toughness
// become 4/4" that lasts until end of turn, so counters and later +N/+N
// effects still apply on top of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "013a465b-7034-45fe-bf62-c3d273131ec5",
		Name:         "Square Up",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, item.Targets[0].ID, nil,
				"Square Up — base power and toughness 4/4 until end of turn",
				game.SetBasePTMods(4, 4)...)
		},
	})
}
