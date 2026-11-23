package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "02520741-d7f0-46e6-acbe-010b65fc24c0",
		Name:            "Angel of Despair",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Angel of Despair — destroy target permanent", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target permanent")),
		},
	})
}
