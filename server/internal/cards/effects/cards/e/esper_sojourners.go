package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Esper Sojourners — Artifact Creature — Vedalken Wizard {W}{U}{B}, 2/3
// (Alara Reborn, #5586):
//
//	"When you cycle this card and when this creature dies, you may tap or
//	 untap target permanent.
//	 Cycling {2}{U}"
//
// One ability with two conditions (S82OldB501CycledOrDies). "You may tap or
// untap" is Janjeet Sentry's reading: a pick made as the trigger resolves,
// and tapping a tapped permanent or untapping an untapped one does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c59db70-9cb4-4edf-ba6e-12ffa7c06b7b",
		Name:         "Esper Sojourners",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.S82OldB501CycledOrDies("Esper Sojourners — you may tap or untap target permanent",
					effects.B38TapOrUntapTarget("Esper Sojourners")),
				effects.TargetPermanent("target permanent")),
		},
	})
}
