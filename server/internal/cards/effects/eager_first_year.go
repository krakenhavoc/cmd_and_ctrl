package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "74788d52-dad7-4c10-8567-a2c86ed90a67",
		Name:         "Eager First-Year",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Magecraft("Eager First-Year — gets +1/+0 until end of turn",
				thisGetsUntilEndOfTurn(1, 0, "Eager First-Year — magecraft pump")),
		},
	})
}
