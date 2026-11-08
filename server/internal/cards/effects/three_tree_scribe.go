package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Three Tree Scribe — Creature — Frog Druid {1}{G}, 2/3 (Bloomburrow,
// #4467):
//
//	"Whenever this creature or another creature you control leaves the
//	 battlefield without dying, put a +1/+1 counter on target creature
//	 you control."
//
// Aang, Airbending Master's condition (creatureYouControlLeftWithoutDying,
// CR 700.4): a creature you controlled went anywhere but a graveyard, a
// bounce, a blink or an exile, judged by its type and controller as it
// left. The Scribe counts itself, off the CR 603.10a look-back. It is
// not "one or more", so each creature that leaves is its own trigger
// (CR 603.2c). The counter goes on the still-legal target
// (PlusOneCounterOnTheTarget).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d2a7741-1dff-4fac-92d8-abea24f23f97",
		Name:         "Three Tree Scribe",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return creatureYouControlLeftWithoutDying(ev, source, g)
			}, "Three Tree Scribe — put a +1/+1 counter on target creature you control", PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
