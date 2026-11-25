package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Deadeye Rig-Hauler — Creature — Human Pirate {3}{U}, 3/2 (Rivals of Ixalan,
// S82 OLD batch 218, #5531):
//
//	"Raid — When this creature enters, if you attacked this turn, you may return
//	 target creature to its owner's hand."
//
// Gorehorn Raider's raid enters trigger with Jeskai Barricade's optional bounce:
// raid is an intervening "if" (CR 603.4), checked as it triggers and again as it
// resolves; the target (any creature, this one included) is chosen as the
// trigger goes on the stack and re-checked as it resolves (CR 608.2b), and the
// "may" is answered as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1355424e-9909-41dc-8d43-b8d21b0919bf",
		Name:         "Deadeye Rig-Hauler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.TmcB6YouAttackedThisTurn),
					"Deadeye Rig-Hauler — Raid: return target creature to its owner's hand",
					effects.Do(effects.If{Cond: effects.TmcB6CondYouAttacked{}, Then: []effects.Applier{
						effects.Instruction(func(ctx *effects.Context) error { return effects.BounceChosenTarget(ctx.Game, ctx.Item) }),
					}})),
				"Deadeye Rig-Hauler — return target creature to its owner's hand?"),
				effects.TargetCreature("target creature")),
		},
	})
}
