package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Parasitic Strix — Artifact Creature — Bird {2}{U}, 2/2 (Conflux,
// #5587):
//
//	"Flying
//	 When this creature enters, if you control a black permanent, target
//	 player loses 2 life and you gain 2 life."
//
// An intervening-if enters trigger (CR 603.4, conB2WhenThisEntersIfYouControlA)
// whose body is Kalastria Highborn's drain. The Strix itself is blue, so
// it never counts as the black permanent unless something has made it
// black.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "975b4e0f-1133-4ffa-a6b1-d319d248dcde",
		Name:            "Parasitic Strix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.ConB2WhenThisEntersIfYouControlA("B", "black",
				"Parasitic Strix — target player loses 2 life and you gain 2 life",
				effects.TargetPlayer("target player"),
				effects.Do(effects.TargetPlayerLosesLife{N: 2}, effects.GainLife{Amount: 2})),
		},
	})
}
