package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subterranean Shambler — Creature — Elemental, {3}{R}, 2/3:
//
//	"Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters or leaves the battlefield, it deals 1 damage to each creature without flying."
//
// "Enters or leaves the battlefield" is one ability with two trigger
// conditions; the damage goes to each creature without flying as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d06be0a6-0d45-4e95-b132-88ce80dd7cb6",
		Name:         "Subterranean Shambler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Echo("Subterranean Shambler", "{3}{R}"),
			WhenThisEntersOrLeaves(
				"Subterranean Shambler — 1 damage to each creature without flying",
				func(g *game.Game, item *game.StackItem) error {
					return damageEachMatching(NewContext(g, item), And(Creature(), WithoutKeyword("flying")), 1)
				}),
		},
	})
}
