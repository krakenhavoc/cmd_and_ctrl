package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Angel of Despair — Creature — Angel {3}{W}{W}{B}{B}, 5/5 (Guildpact,
// #5599):
//
//	"Flying
//	 When this creature enters, destroy target permanent."
//
// Spine of Ish Sah's enters trigger: any permanent, the Angel itself
// included, chosen as the trigger goes on the stack and checked again
// as it resolves (CR 608.2b). Regeneration and indestructible apply as
// usual.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "02520741-d7f0-46e6-acbe-010b65fc24c0",
		Name:            "Angel of Despair",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Angel of Despair — destroy target permanent", destroyChosenPermanent),
				TargetPermanent("target permanent")),
		},
	})
}
