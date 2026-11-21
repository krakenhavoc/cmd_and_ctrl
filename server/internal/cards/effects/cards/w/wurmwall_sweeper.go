package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wurmwall Sweeper — Artifact — Spacecraft {2}, 2/2 (Edge of Eternities,
// #4461):
//
//	"When this Spacecraft enters, surveil 2.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 4+.)
//	 4+ | Flying"
//
// Pinnacle Kill-Ship's station lines (Station, SpacecraftAt and
// ThresholdKeywords, CR 702.184a and 721.2a) with Starbreach Whale's
// entry surveil.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "320ed2ec-ef06-41fb-ac8a-72579f11544c",
		Name:         "Wurmwall Sweeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wurmwall Sweeper — surveil 2", effects.Do(effects.Surveil{N: 2})),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(4, 2, 2),
			effects.ThresholdKeywords(4, "flying"),
		),
	})
}
