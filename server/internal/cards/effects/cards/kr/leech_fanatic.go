package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "17bcd92b-1826-4a36-9048-49e5e65a12fd",
		Name:         "Leech Fanatic",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "lifelink"),
		},
	})
}
