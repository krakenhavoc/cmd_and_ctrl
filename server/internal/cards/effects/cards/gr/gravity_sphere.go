package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravity Sphere — World Enchantment {2}{R}:
//
//	"All creatures lose flying."
//
// A layer-6 removal over every creature on the battlefield
// (LoseKeywords, lose_keywords.go). It is "lose", not "can't have", so
// a flying grant with a later timestamp gives the keyword back (CR
// 613.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ddf93fe-980b-4dc4-b56f-6a2ee50100a6",
		Name:         "Gravity Sphere",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.LoseKeywords(effects.AllCreatures, "flying")},
	})
}
