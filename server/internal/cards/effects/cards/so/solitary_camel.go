package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Solitary Camel — {2}{W} Creature — Camel 3/2 (Hour of Devastation):
//
//	"This creature has lifelink as long as you control a Desert or there is a
//	 Desert card in your graveyard."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2eac0c6a-b289-4b75-9571-59869e62d180",
		Name:         "Solitary Camel",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.OldB239YouHaveADesert(g, source.Controller)
			}, "lifelink"),
		},
	})
}
