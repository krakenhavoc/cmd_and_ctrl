package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guul Draz Overseer — Creature — Vampire {4}{B}{B}, 3/4 (Battle for
// Zendikar):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, other creatures you
//	 control get +1/+0 until end of turn. If that land is a Swamp, those
//	 creatures get +2/+0 until end of turn instead."
//
// The creatures you control as the ability resolves are the ones pumped
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9eff1ce5-41e4-4ae6-9480-2b549c8546bd",
		Name:            "Guul Draz Overseer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Guul Draz Overseer — other creatures you control get +1/+0 until end of turn, +2/+0 if that land is a Swamp",
				func(g *game.Game, item *game.StackItem) error {
					power := 1
					if effects.S82OldB309EnteredLandIs(g, item, "Swamp") {
						power = 2
					}
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.NotSelf(ctx.Source())), Power: power,
						Label: "Guul Draz Overseer — other creatures get +P/+0",
					}.Apply(ctx)
				}),
		},
	})
}
