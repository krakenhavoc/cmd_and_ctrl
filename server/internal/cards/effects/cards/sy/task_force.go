package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Task Force — Creature — Human Rebel {2}{W}, 1/3 (Jumpstart 2022):
//
//	"Whenever this creature becomes the target of a spell or ability, it gets
//	 +0/+3 until end of turn."
//
// Cephalid Illusionist's trigger (the engine's EventBecomesTarget for this
// creature, any controller's spell or ability) with a +0/+3 until end of turn
// on the source. One spell naming it twice triggers twice (CR 115.3), as
// printed; a Task Force that left gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b80415cd-5887-4146-852e-73119fcd3ec6",
		Name:         "Task Force",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.Self, "Task Force — it gets +0/+3 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.OnBattlefield(g, ctx.Source()) || effects.SourceIsNewObject(g, item) {
						return nil
					}
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 0, Toughness: 3,
						Label: "Task Force — +0/+3 until end of turn"}.Apply(ctx)
				}),
		},
	})
}
