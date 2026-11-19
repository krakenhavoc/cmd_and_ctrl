package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kapsho Kitefins — Creature — Fish {4}{U}{U}, 3/3:
//
//	"Flying
//	 Whenever this creature or another creature you control enters, tap
//	 target creature an opponent controls."
//
// One trigger on CreatureEnteredUnderYourControl (the source included,
// Impact Tremors' predicate), with a real target clause chosen as the
// trigger goes on the stack.
//
// No simplification.
func init() {
	const label = "Kapsho Kitefins — tap target creature an opponent controls"
	Register(Spec{
		OracleID:        "b9b205db-068e-44c7-9169-5d8f13d4b73a",
		Name:            "Kapsho Kitefins",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventETB, CreatureEnteredUnderYourControl, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return TapTarget{Target: t.ID}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
