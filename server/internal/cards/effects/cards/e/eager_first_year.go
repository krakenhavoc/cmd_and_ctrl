package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eager First-Year — Creature — Human Wizard {1}{W}, 2/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Magecraft — Whenever you cast or copy an instant or sorcery spell,
//	 this creature gets +1/+0 until end of turn."
//
// Veyran, Voice of Duality's magecraft pump at +1/+0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74788d52-dad7-4c10-8567-a2c86ed90a67",
		Name:         "Eager First-Year",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Magecraft("Eager First-Year — gets +1/+0 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 0, "Eager First-Year — magecraft pump")),
		},
	})
}
