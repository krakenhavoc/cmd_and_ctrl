package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jadecraft Artisan — Creature — Merfolk Shaman {3}{G}, 3/3 (Rivals of
// Ixalan, #5531):
//
//	"When this creature enters, target creature gets +2/+2 until end of
//	 turn."
//
// Yeva's Forcemage's entry: the target is chosen as the trigger goes on
// the stack (CR 603.3d) and re-checked as it resolves (CR 608.2b); it may
// be the Artisan itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9290f549-3e5d-4e39-a20e-8be6b674abc7",
		Name:         "Jadecraft Artisan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Jadecraft Artisan — target creature gets +2/+2 until end of turn",
				effects.BoostTargetUntilEOT("Jadecraft Artisan — +2/+2", 2, 2)), effects.TargetCreature("target creature")),
		},
	})
}
