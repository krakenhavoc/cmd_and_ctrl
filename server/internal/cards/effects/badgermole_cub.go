package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Badgermole Cub — Creature — Badger Mole {1}{G}, 2/2 (Avatar: The Last
// Airbender, #4459):
//
//	"When this creature enters, earthbend 1. (Target land you control
//	 becomes a 0/0 creature with haste that's still a land. Put a +1/+1
//	 counter on it. When it dies or is exiled, return it to the
//	 battlefield tapped.)
//	 Whenever you tap a creature for mana, add an additional {G}."
//
// The entry is the earthbend keyword action (earthbend.go) with its own
// target clause, chosen as the trigger goes on the stack.
//
// The second ability is a triggered mana ability (CR 605.1b, the ruling):
// it resolves at once, with no stack (CR 605.4a). The engine fires it
// only for a mana ability with {T} in its cost (ManaProduced comes from a
// permanent TAPPED for mana), so a "tap an untapped creature you
// control" cost and a creature tapped to pay another object's ability
// don't trigger it, as the ruling says. "A creature" is read through the
// layers on the tapped permanent, so a land an earthbend animated counts
// while it is a creature; "you tap" is the player whose pool the mana
// went to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b0afb89-0944-4861-b9c3-e909e2ac215e",
		Name:         "Badgermole Cub",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			avB14WhenEntersEarthbend("Badgermole Cub — earthbend 1", 1),
		},
		ManaTriggers: []game.ManaTrigger{{
			Label: "Badgermole Cub — add an additional {G}",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, _ *game.Game) bool {
				return prod.Source.IsCreature() && prod.Controller == source.Controller
			},
			Produced: func(game.ManaProduced, *game.Card, *game.Game) string { return "{G}" },
		}},
	})
}
