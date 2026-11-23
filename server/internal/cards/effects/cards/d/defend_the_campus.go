package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defend the Campus — Instant {3}{W} (Foundations Jumpstart, #5929):
//
//	"Choose one —
//	 • Creatures you control get +1/+1 until end of turn.
//	 • Destroy target creature with power 4 or greater."
//
// Only the second bullet targets, so the picker asks for a creature with
// power 4 or greater only when that mode is chosen; the power is checked
// again as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "274c2781-f40a-49b4-a039-669739beaab6",
		Name:         "Defend the Campus",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: effects.ChooseOne(
			effects.Mode("Creatures you control get +1/+1 until end of turn."),
			effects.Mode("Destroy target creature with power 4 or greater.",
				effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) {
				return effects.BoostUntilEOT{
					Match:     effects.And(effects.Creature(), effects.YouControl()),
					Power:     1,
					Toughness: 1,
					Label:     "Defend the Campus — +1/+1",
				}.Apply(ctx)
			}
			if ctx.HasMode(1) && len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
