package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Cat — Creature — Zombie Cat {1}{B}, 1/1 (Dark Ascension, #5571):
//
//	"When this creature dies, target opponent discards a card at
//	 random."
//
// The discard is the engine's random one (DiscardCards), drawn from the
// keyed random stream; an empty hand discards nothing. An opponent who has
// left the game in response is an illegal target and the trigger does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "10edf7e0-d6be-4510-819f-8834bedfba41",
		Name:         "Black Cat",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Black Cat — target opponent discards a card at random",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return effects.DiscardCards{Player: t.ID, N: 1}.Apply(ctx)
						}
					}
					return nil
				}))),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
