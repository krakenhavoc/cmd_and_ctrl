package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "b9b205db-068e-44c7-9169-5d8f13d4b73a",
		Name:            "Kapsho Kitefins",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl, label,
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return effects.TapTarget{Target: t.ID}.Apply(ctx)
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
