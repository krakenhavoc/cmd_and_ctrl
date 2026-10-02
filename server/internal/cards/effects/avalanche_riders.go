package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Avalanche Riders — Creature — Human Nomad, {3}{R}, 2/2:
//
//	"Haste
//	 Echo {3}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, destroy target land."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c9048d35-aff7-45ae-ae97-db89397b3e74",
		Name:            "Avalanche Riders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Echo("Avalanche Riders", "{3}{R}"),
			Targeting(WhenThisEnters("Avalanche Riders — destroy target land", destroyFirstLegalCardTarget),
				TargetPermanent("target land", Land())),
		},
	})
}
