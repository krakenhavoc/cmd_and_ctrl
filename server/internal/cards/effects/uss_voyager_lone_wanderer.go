package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// U.S.S. Voyager, Lone Wanderer — Legendary Artifact — Spacecraft {3}{G},
// 3/3 (Star Trek, #4973):
//
//	"When U.S.S. Voyager enters, reveal cards from the top of your library
//	 until you reveal two land cards. Put those land cards onto the
//	 battlefield tapped and the rest on the bottom of your library in a
//	 random order.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 8+.)
//	 8+ | Flying, vigilance"
//
// The enters trigger is Open the Way's reveal with X fixed at two
// (openTheWayReveal): the run stops on the second land card, the lands
// enter tapped, and a library with fewer than two land cards reveals
// itself whole and puts the lands it has. The station frame is
// Galvanizing Sawship's (Station, SpacecraftAt, ThresholdKeywords).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33e8904a-9285-4031-900c-0a5c8f0bf6c0",
		Name:         "U.S.S. Voyager, Lone Wanderer",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("U.S.S. Voyager, Lone Wanderer — reveal until two land cards; put them onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return openTheWayReveal(NewContext(g, item), 2)
				}),
		},
		Activated: []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(8, 3, 3),
			ThresholdKeywords(8, "flying", "vigilance"),
		),
	})
}
