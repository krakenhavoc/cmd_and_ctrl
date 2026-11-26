package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acid Web Spider — Creature — Spider {3}{G}{G}, 3/5 (Scars of Mirrodin, S82
// OLD batch 471, #5577):
//
//	"Reach
//	 When this creature enters, you may destroy target Equipment."
//
// Batterhorn's optional enter-destroy aimed at an Equipment: the target is
// chosen as the trigger goes on the stack and the "may" is asked on resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23d4f436-a417-4a26-b5b8-c698f11a186b",
		Name:            "Acid Web Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.WhenThisEnters("Acid Web Spider — you may destroy target Equipment", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target Equipment", effects.HasSubtype("Equipment"))),
				"Acid Web Spider — destroy target Equipment?"),
		},
	})
}
