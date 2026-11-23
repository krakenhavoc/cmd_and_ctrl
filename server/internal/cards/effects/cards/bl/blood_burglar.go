package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blood Burglar — Creature — Vampire Rogue {1}{B}, 2/2 (Foundations
// Jumpstart):
//
//	"During your turn, this creature has lifelink."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c975fda-471a-4d41-9bb6-ea42edc6a0e6",
		Name:         "Blood Burglar",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "lifelink"),
		},
	})
}
