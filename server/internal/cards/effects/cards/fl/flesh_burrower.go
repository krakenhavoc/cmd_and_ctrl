package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flesh Burrower — Creature — Insect {1}{G}, 2/2 (Duskmourn: House of
// Horror, #4466):
//
//	"Deathtouch
//	 Whenever this creature attacks, another target creature you control
//	 gains deathtouch until end of turn."
//
// Foot Elite's "another target creature you control" attack trigger with
// the keyword alone: the shared pumpAndGrantFirstTarget body gives the
// still-legal target (CR 608.2b) deathtouch and nothing else, and nothing
// at all if it left. The target need not be attacking.
//
// No simplification.
func init() {
	const label = "Flesh Burrower — another target creature you control gains deathtouch until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "04c4250e-014a-4067-80e8-4134069354f7",
		Name:            "Flesh Burrower",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks(label, effects.PumpAndGrantFirstTarget(0, 0, label, "deathtouch")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
