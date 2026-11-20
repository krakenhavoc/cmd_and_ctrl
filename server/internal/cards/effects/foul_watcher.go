package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foul Watcher — Creature — Nightmare Bird {1}{U}, 1/2 (Modern Horizons 2,
// #5501):
//
//	"Flying
//	 When this creature enters, surveil 1. (Look at the top card of your
//	 library. You may put that card into your graveyard.)
//	 Delirium — This creature gets +1/+0 as long as there are four or
//	 more card types among cards in your graveyard."
//
// Shore Lurker's enter trigger and Grim Flayer's delirium static
// (whileYouHave over drcDeliriumMet), here +1/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b50dd71f-3664-41d2-be06-57729064a591",
		Name:            "Foul Watcher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Foul Watcher — surveil 1", Do(Surveil{N: 1})),
		},
		Static: []game.StaticAbility{{
			Label:     "Delirium — This creature gets +1/+0 as long as there are four or more card types among cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileYouHave(drcDeliriumMet, selfCreature),
			Apply:     pumpBy(1, 0),
		}},
	})
}
