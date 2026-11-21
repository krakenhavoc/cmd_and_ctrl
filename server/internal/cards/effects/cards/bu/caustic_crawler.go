package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caustic Crawler — Creature — Insect {3}{B}{B}, 4/3 (Worldwake, #5582):
//
//	"Landfall — Whenever a land you control enters, you may have target
//	 creature get -1/-1 until end of turn."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d)
// and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91c83b1e-32d9-4bb4-9a71-05fbb0cb8134",
		Name:         "Caustic Crawler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.Landfall("Caustic Crawler — target creature gets -1/-1 until end of turn (landfall)",
					effects.NeoB07TargetGetsUntilEOT("Caustic Crawler", -1, -1)),
				effects.TargetCreature("target creature")),
				"Caustic Crawler — have target creature get -1/-1 until end of turn?"),
		},
	})
}
