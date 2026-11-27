package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Surging Dementia — Sorcery {1}{B} (Coldsnap, #5597):
//
//	"Ripple 4 (...)
//	 Target player discards a card."
//
// Mind Peel's discard (the target player chooses the card) with ripple
// (CR 702.60a) as a trigger of the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd7f46c2-98f7-46e6-92ce-51fa95d2f0c8",
		Name:         "Surging Dementia",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Triggered:    []game.TriggeredAbility{effects.S82OldB566Ripple("Surging Dementia", 4)},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetOpponentDiscards(1)(ctx.Game, item)
		},
	})
}
