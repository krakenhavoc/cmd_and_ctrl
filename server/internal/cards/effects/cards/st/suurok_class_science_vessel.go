package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "5d2016ed-3ea0-4938-b95e-ee582307e475",
		Name:         "Suurok-Class Science Vessel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Suurok-Class Science Vessel — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(5, 3, 2),
			effects.ThresholdKeywords(5, "flying"),
		),
	})
}
