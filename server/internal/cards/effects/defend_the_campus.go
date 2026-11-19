package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "274c2781-f40a-49b4-a039-669739beaab6",
		Name:         "Defend the Campus",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			Mode("Creatures you control get +1/+1 until end of turn."),
			Mode("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4))),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				return BoostUntilEOT{
					Match:     And(Creature(), YouControl()),
					Power:     1,
					Toughness: 1,
					Label:     "Defend the Campus — +1/+1",
				}.Apply(ctx)
			}
			if ctx.HasMode(1) && len(item.Targets) > 0 && item.Targets[0].Kind == game.TargetCard {
				return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
			}
			return nil
		},
	})
}
