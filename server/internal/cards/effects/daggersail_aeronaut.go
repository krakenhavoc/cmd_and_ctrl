package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daggersail Aeronaut — Creature — Goblin {3}{R}, 3/2 (Core Set 2020, #5518):
//
//	"During your turn, this creature has flying."
//
// Hookblade Veteran's grant: a layer 6 keyword on itself, on only while
// its controller is the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c0ef3b7-b368-4a5b-8cbe-bf1f47610f74",
		Name:         "Daggersail Aeronaut",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
	})
}
