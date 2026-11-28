package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kavu Runner — Creature — Kavu {3}{R}, 3/3 (Invasion, S82 older sets batch 664, #5616):
//
//	"This creature has haste as long as no opponent controls a white or blue
//	 creature."
//
// Mobile Homestead's haste on itself, read live on every layer pass: the Runner
// has haste only while no creature that is white or blue, as the layers have it,
// is controlled by an opponent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c90b7fa-d5bf-4c46-92d5-3014ea1bec50",
		Name:         "Kavu Runner",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID &&
				effects.OldB664NoOpponentControlsACreatureOfColor(g, source, "W", "U")
		}, "haste")},
	})
}
