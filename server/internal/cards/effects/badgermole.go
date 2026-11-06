package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Badgermole — Creature — Badger Mole {4}{G}, 4/4 (Avatar: The Last
// Airbender, #4459):
//
//	"When this creature enters, earthbend 2. (Target land you control
//	 becomes a 0/0 creature with haste that's still a land. Put two
//	 +1/+1 counters on it. When it dies or is exiled, return it to the
//	 battlefield tapped.)
//	 Creatures you control with +1/+1 counters on them have trample."
//
// The entry trigger is the earthbend keyword action with the shared
// target clause (earthbend.go); a land that is no longer a legal target
// on resolution is not earthbent (the fourth ruling). The trample is
// Emil, Vastlands Roamer's layer-6 grant, read live: a creature that
// gets its first +1/+1 counter later has trample from then on, the
// earthbent land among them, and one that loses its last counter loses
// it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2084a7f6-226b-4fb4-a30e-703f674d33b4",
		Name:         "Badgermole",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16GrantKeywords(b42CreatureYouControlWithAPlusOneCounter, "trample")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Badgermole — earthbend 2", func(g *game.Game, item *game.StackItem) error {
					return EarthbendFirstTarget(EarthbendCount(2))(item, NewContext(g, item))
				}),
				EarthbendTargets(),
			),
		},
	})
}
