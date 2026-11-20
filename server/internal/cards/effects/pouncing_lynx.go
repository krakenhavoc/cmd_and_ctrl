package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pouncing Lynx — Creature — Cat {1}{W}, 2/1 (War of the Spark, #5520):
//
//	"During your turn, this creature has first strike."
//
// Fresh-Faced Recruit's static: first strike on itself while its
// controller is the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8d92f97-d54c-4954-8324-b9c0a9b96ec2",
		Name:         "Pouncing Lynx",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "first strike"),
		},
	})
}
