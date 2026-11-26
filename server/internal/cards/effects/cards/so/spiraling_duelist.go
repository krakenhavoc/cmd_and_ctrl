package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spiraling Duelist — Creature — Human Berserker {2}{R}{R}, 3/1 (Mirrodin
// Besieged, S82 OLD b465, #5576):
//
//	"Metalcraft — This creature has double strike as long as you control three
//	 or more artifacts."
//
// The metalcraft self grant (SomB2MetalcraftKeywords), in layer 6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cc803d43-cbbe-4964-8077-ba405a741047",
		Name:         "Spiraling Duelist",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.SomB2MetalcraftKeywords("Metalcraft — This creature has double strike as long as you control three or more artifacts.", "double strike"),
		},
	})
}
