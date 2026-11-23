package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crippling Chill — Instant {2}{U} (Jumpstart 2022, #5948):
//
//	"Tap target creature. It doesn't untap during its controller's next untap
//	 step.
//	 Draw a card."
//
// Snow Day's tap-and-hold on one target (the hold is the creature's
// controller's next untap step), then a draw for the caster. A target that
// is illegal as it resolves leaves the whole spell without effect, the draw
// included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4459117a-d11c-44f4-9aac-caefc5b2d6f2",
		Name:         "Crippling Chill",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    effects.GapB172CripplingChill,
	})
}
