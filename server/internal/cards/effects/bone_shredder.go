package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Shredder — Creature — Phyrexian Minion, {2}{B}, 1/1:
//
//	"Flying
//	 Echo {2}{B} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 When this creature enters, destroy target nonartifact, nonblack creature."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69aa1ac7-17da-4fc9-a3dc-df8a52841f4a",
		Name:            "Bone Shredder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Echo("Bone Shredder", "{2}{B}"),
			Targeting(WhenThisEnters("Bone Shredder — destroy target nonartifact, nonblack creature", destroyFirstLegalCardTarget),
				TargetCreature("target nonartifact, nonblack creature", Not(Artifact()), NonBlack())),
		},
	})
}
