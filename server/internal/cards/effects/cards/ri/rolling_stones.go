package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rolling Stones — Enchantment {1}{W} (Eighth Edition, #5607):
//
//	"Wall creatures can attack as though they didn't have defender."
//
// Every Wall on the battlefield, whoever controls it (CR 508.1a, 702.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7ce533d0-8b6e-46d8-b21e-a2932bdffb67",
		Name:         "Rolling Stones",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.DefenderWaiverStatic("Wall creatures can attack as though they didn't have defender.",
				func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.HasSubtype(source.TextSubtype("Wall"))
				}),
		},
	})
}
