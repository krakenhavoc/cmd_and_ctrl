package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Saru, Unthreatened — {2}{G} Legendary Creature — Kelpien Officer 3/2
// (Star Trek, #4973):
//
//	"Flash
//	 When Saru enters, another target creature you control gains hexproof
//	 until end of turn."
//
// "Another" keeps Saru himself out of the target set; a target that left
// or changed control in response is illegal and the ability does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6a21e53-6f62-4874-b9cc-84ae177a0da6",
		Name:            "Saru, Unthreatened",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Saru, Unthreatened — another target creature you control gains hexproof until end of turn",
				effects.EbTargetCreatureGainsUntilEOT("hexproof", "Saru, Unthreatened — hexproof until end of turn")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
		},
	})
}
