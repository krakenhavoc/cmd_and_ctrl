package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watcher of the Wayside — Artifact Creature — Golem {3}, 3/2 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, target player mills two cards. You gain 2
//	 life."
//
// Desperate Bloodseeker's targeted mill (targetPlayerMills, CR 701.17a)
// followed by the life gain. The player is chosen as the trigger goes on
// the stack and checked again as it resolves (CR 608.2b); you may name
// yourself. If that player has become an illegal target, the trigger does
// nothing at all, the life gain included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b30eafed-2755-4423-9592-c8b83c68804b",
		Name:         "Watcher of the Wayside",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Watcher of the Wayside — target player mills two cards; you gain 2 life", Do(
				Instruction(func(ctx *Context) error { return targetPlayerMills(2)(ctx.Game, ctx.Item) }),
				GainLife{Amount: 2},
			)), TargetPlayer("target player")),
		},
	})
}
