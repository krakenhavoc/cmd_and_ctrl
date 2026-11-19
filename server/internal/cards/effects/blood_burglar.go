package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blood Burglar — Creature — Vampire Rogue {1}{B}, 2/2 (Foundations
// Jumpstart):
//
//	"During your turn, this creature has lifelink."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c975fda-471a-4d41-9bb6-ea42edc6a0e6",
		Name:         "Blood Burglar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "lifelink"),
		},
	})
}
