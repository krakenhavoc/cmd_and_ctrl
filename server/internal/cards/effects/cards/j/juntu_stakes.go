package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Juntu Stakes — Artifact {2} (Invasion):
//
//	"Creatures with power 1 or less don't untap during their controllers'
//	 untap steps."
//
// Back to Basics's untap restriction (CR 502.3), asked at each untap step of
// the creature's controller, with the power read as the untap step begins (the
// creature's effective power after layers and counters; a creature already
// untapped is unaffected).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efd73b8a-1caa-4fc8-855a-cad33ea54097",
		Name:         "Juntu Stakes",
		Completeness: effects.CompletenessFull,
		UntapStepRestrictions: []game.UntapStepRestriction{
			effects.DoesntUntapDuringTheirControllersUntapSteps(func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
				return c.IsCreature() && c.CurrentPower() <= 1
			}),
		},
	})
}
