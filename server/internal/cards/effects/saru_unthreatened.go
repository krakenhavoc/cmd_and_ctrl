package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b6a21e53-6f62-4874-b9cc-84ae177a0da6",
		Name:            "Saru, Unthreatened",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Saru, Unthreatened — another target creature you control gains hexproof until end of turn",
				ebTargetCreatureGainsUntilEOT("hexproof", "Saru, Unthreatened — hexproof until end of turn")),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
