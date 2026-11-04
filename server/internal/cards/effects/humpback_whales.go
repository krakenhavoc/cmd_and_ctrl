package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Humpback Whales — Creature — Whale {4}{U}{U}, 5/6 (Star Trek, #4973):
//
//	"Vigilance, ward {2}
//	 Basic landcycling {2} ({2}, Discard this card: Search your library
//	 for a basic land card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// Vigilance is a printed keyword, ward {2} the shared Ward trigger
// (CR 702.21a) and basic landcycling the shared hand ability
// (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7285536-6cc1-448e-839a-606d585fbd9d",
		Name:            "Humpback Whales",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{2}"), "Humpback Whales — ward {2}"),
		},
		Activated: []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
