package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "10edf7e0-d6be-4510-819f-8834bedfba41",
		Name:         "Black Cat",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Black Cat — target opponent discards a card at random",
				Do(Instruction(func(ctx *Context) error {
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return DiscardCards{Player: t.ID, N: 1}.Apply(ctx)
						}
					}
					return nil
				}))),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
