package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prying Questions — Sorcery {2}{B} (Eldritch Moon, S82 OLD, #5541):
//
//	"Target opponent loses 3 life and puts a card from their hand on top of
//	 their library."
//
// The life loss, then the opponent's own pick of one card from their hand
// (Brainstorm's put-back, for the targeted player; an empty hand is not
// asked).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f33ef275-632e-46c8-a735-042dee340a12",
		Name:         "Prying Questions",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, LifeLoss: 3}),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GainLife{Player: effects.TargetedPlayer(ctx), Amount: -3}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.PutFromHandOnTopInAnyOrder{Player: effects.TargetedPlayer(ctx), N: 1,
					Label: "Prying Questions — put a card from your hand on top of your library"}.Apply(ctx)
			}),
		),
	})
}
