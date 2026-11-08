package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "320ed2ec-ef06-41fb-ac8a-72579f11544c",
		Name:         "Wurmwall Sweeper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wurmwall Sweeper — surveil 2", Do(Surveil{N: 2})),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(4, 2, 2),
			ThresholdKeywords(4, "flying"),
		),
	})
}
