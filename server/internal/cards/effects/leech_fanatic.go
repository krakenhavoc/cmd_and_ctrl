package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leech Fanatic — Creature — Human Warlock {1}{B}, 2/2 (Strixhaven: School
// of Mages, #5502):
//
//	"During your turn, this creature has lifelink."
//
// Blood Burglar's layer-6 grant, live only while its controller is the
// active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17bcd92b-1826-4a36-9048-49e5e65a12fd",
		Name:         "Leech Fanatic",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			}, "lifelink"),
		},
	})
}
