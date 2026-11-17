package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phalanx Vanguard — Creature — Human Soldier {1}{W}, 2/2 (The Brothers'
// War, #5482):
//
//	"Vigilance
//	 Whenever an artifact you control enters, this creature gets +1/+0
//	 until end of turn."
//
// ArtifactEnteredUnderYourControl counts the Vanguard too, should it ever
// be an artifact. Each artifact is its own trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "265f461e-d076-4119-8ea4-96360305f3a0",
		Name:            "Phalanx Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{s82BroB07PumpSelfOnArtifact("Phalanx Vanguard")},
	})
}
