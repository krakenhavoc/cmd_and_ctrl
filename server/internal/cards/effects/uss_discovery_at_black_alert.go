package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Discovery, at Black Alert — {3}{G}{G} Legendary Artifact —
// Spacecraft 5/5 (Star Trek, #4973):
//
//	"Whenever U.S.S. Discovery enters or attacks, discover 5.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Flying"
//
// One ability, two conditions (Sun Titan's WhenThisEntersOrAttacks); it
// can only attack once it is a creature, at six or more charge counters.
// Discover is CR 701.57.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13dbfe3f-1e04-4c4f-953c-7c949dba9d33",
		Name:         "U.S.S. Discovery, at Black Alert",
		Purpose:      game.Purpose{Discover: 5},
		Completeness: CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("U.S.S. Discovery — discover 5", DiscoverN(5)),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 5, 5),
			ThresholdKeywords(6, "flying"),
		),
	})
}
