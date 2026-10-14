package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Dead — Creature — Zombie {3}{B}, 3/2:
//
//	"Whenever this creature or another Zombie dies, each opponent loses
//	 1 life."
//
// "Another Zombie" is anyone's: an opponent's Zombie dying drains the
// table too. Whether the creature that died was a Zombie is its
// last-known subtypes (leftAsSubtype, CR 603.10a), so a changeling
// counts. The Vengeful Dead itself counts whatever its types were, as
// "this creature" says; it sees its own death by the leaves-the-
// battlefield look-back (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "277c8ee9-0157-4e45-96ad-1b67716955ee",
		Name:         "Vengeful Dead",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventLTB, thisOrAnotherZombieDied,
				"Vengeful Dead — each opponent loses 1 life", Do(EachOpponentLosesLife{N: 1})),
				game.Purpose{DeathPayoff: true}),
		},
	})
}

// thisOrAnotherZombieDied is "this creature or another Zombie dies".
func thisOrAnotherZombieDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	return ev.CardID == source.InstanceID || leftAsSubtype(ev, dead, "Zombie")
}
