package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Enterprise-E, Sovereign — Legendary Artifact — Spacecraft
// {3}{W}{W}, 7/7 (Star Trek Commander, #4974):
//
//	"When U.S.S. Enterprise-E enters, destroy all tapped creatures.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 8+.)
//	 8+ | Flying, vigilance"
//
// The entry trigger is Don't Move's sweep (every tapped creature, whoever
// controls it, at once), and the station body is The Seriema's with
// flying and vigilance above eight counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa667031-25cf-45c4-97ff-19a73cb30018",
		Name:         "U.S.S. Enterprise-E, Sovereign",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("U.S.S. Enterprise-E — destroy all tapped creatures",
				Do(DestroyAllMatching{Match: dontMoveTapped})),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 7, 7),
			ThresholdKeywords(8, "flying", "vigilance"),
		),
	})
}
