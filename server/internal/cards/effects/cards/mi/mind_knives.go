package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mind Knives — Sorcery {1}{B} (Portal, S82 OLD batch 781, #5630):
//
//	"Target opponent discards a card at random."
//
// Black Cat's discard: the engine's random one (DiscardCards), drawn from the
// keyed random stream; an empty hand discards nothing. An opponent who has left
// the game is an illegal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9be5566-c3df-44cc-9de4-29510c8c245f",
		Name:         "Mind Knives",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.DiscardCards{Player: player, N: 1}.Apply(ctx)
		},
	})
}
