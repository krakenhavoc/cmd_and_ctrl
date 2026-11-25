package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dampening Pulse — Enchantment {3}{U} (Battle for Zendikar):
//
//	"Creatures your opponents control get -1/-0."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe3c81b8-9bca-4760-b677-02bfa7e6137b",
		Name:         "Dampening Pulse",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller != source.Controller
			}, -1, 0),
		},
	})
}
