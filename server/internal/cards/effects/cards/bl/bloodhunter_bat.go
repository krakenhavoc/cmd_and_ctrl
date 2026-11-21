package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodhunter Bat — Creature — Bat {3}{B}, 2/2 (Magic 2013, #5568):
//
//	"Flying
//	 When this creature enters, target player loses 2 life and you gain
//	 2 life."
//
// Blood Artist's drain at two, on an enters trigger: any player may be
// the target, the controller included. The loss is a life change, not
// damage, and the gain is a real life-gain event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7e3e09a2-b45c-4897-ba10-02628cca3f4c",
		Name:            "Bloodhunter Bat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Bloodhunter Bat — target player loses 2 life and you gain 2 life",
					effects.Do(effects.TargetPlayerLosesLife{N: 2}, effects.GainLife{Amount: 2})),
				effects.TargetPlayer("target player")),
		},
	})
}
