package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rubblebelt Boar — Creature — Boar {3}{R}, 3/3 (Guilds of Ravnica, #5525):
//
//	"When this creature enters, target creature gets +2/+0 until end of turn."
//
// Daybreak Charger's targeted enter trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b815cbbd-d79f-4180-8204-504a15df7152",
		Name:         "Rubblebelt Boar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Rubblebelt Boar — target creature gets +2/+0 until end of turn",
				effects.BoostTargetUntilEOT("Rubblebelt Boar — +2/+0", 2, 0)),
				effects.TargetCreature("target creature")),
		},
	})
}
