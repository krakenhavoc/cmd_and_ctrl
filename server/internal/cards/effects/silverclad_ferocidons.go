package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverclad Ferocidons — Creature — Dinosaur {5}{R}{R}, 8/5 (Ixalan,
// #5908):
//
//	"Enrage — Whenever this creature is dealt damage, each opponent
//	 sacrifices a permanent of their choice."
//
// The enrage trigger fires once per batch of damage (CR 120.3, CR 510.2).
// The sacrifice is an edict (CR 701.21): each opponent picks any one of
// their own permanents, in turn order, and none of them has been told
// what the others chose.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3aa5889-0d98-4a8c-a8b1-ada1e9b169e1",
		Name:         "Silverclad Ferocidons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			cmdr2B08EnrageTrigger("Silverclad Ferocidons — Enrage: each opponent sacrifices a permanent of their choice",
				func(g *game.Game, item *game.StackItem) error {
					return EachPlayerSacrifices{
						ExceptController: true,
						Label:            "a permanent",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
