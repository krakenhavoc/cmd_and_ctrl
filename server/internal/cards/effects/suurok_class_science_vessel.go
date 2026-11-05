package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suurok-Class Science Vessel — Artifact — Spacecraft {2}{U}, 3/2
// (Star Trek):
//
//	When this Spacecraft enters, draw a card.
//	Station (Tap another creature you control: Put charge counters
//	equal to its power on this Spacecraft. Station only as a sorcery.
//	It's an artifact creature at 5+.)
//	5+ | Flying
//
// The station seams (Station, SpacecraftAt, ThresholdKeywords; CR 702.184,
// 721) with an ordinary enters draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d2016ed-3ea0-4938-b95e-ee582307e475",
		Name:         "Suurok-Class Science Vessel",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Suurok-Class Science Vessel — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(5, 3, 2),
			ThresholdKeywords(5, "flying"),
		),
	})
}
