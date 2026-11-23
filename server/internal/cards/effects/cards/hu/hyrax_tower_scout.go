package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hyrax Tower Scout — Creature — Human Scout {2}{G}, 3/3 (Theros Beyond Death, #5514):
//
//	"When this creature enters, untap target creature."
//
// The enters trigger untaps any one creature, its controller's or not
// (untapTheTarget, Civic Gardener's body).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1261a72-ccbd-4be9-bc3b-c1bcb807eb85",
		Name:         "Hyrax Tower Scout",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Hyrax Tower Scout — untap target creature", effects.UntapTheTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
