package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "13dbfe3f-1e04-4c4f-953c-7c949dba9d33",
		Name:         "U.S.S. Discovery, at Black Alert",
		Purpose:      game.Purpose{Discover: 5},
		Completeness: effects.CompletenessFull,
		Discovers:    true,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("U.S.S. Discovery — discover 5", effects.DiscoverN(5)),
		},
		Activated: []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(6, 5, 5),
			effects.ThresholdKeywords(6, "flying"),
		),
	})
}
