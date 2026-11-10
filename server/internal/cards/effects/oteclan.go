package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oteclán — Plane — Ixalan (Unknown Event, #4870):
//
//	"When you planeswalk to Oteclán and at the beginning of your upkeep,
//	 chaos ensues.
//	 Whenever chaos ensues, discover 3. (Exile cards from the top of your
//	 library until you exile a nonland card with that mana value or less.
//	 Cast it without paying its mana cost or put it into your hand. Put the
//	 rest on the bottom in a random order.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The first ability is ONE triggered ability with two conditions
// (WhenYouPlaneswalkToThisAndAtYourUpkeep) whose effect is the "chaos
// ensues" instruction (CR 311.7, game.ChaosEnsuesForEffect): it triggers
// every face-up plane's chaos ability, this one's included, as a roll of the
// planar die's chaos symbol would, and is no roll of the die. The chaos
// ability is a discover 3 for the planar controller (CR 701.57).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24b2ba5b-c1cb-4c73-89bf-94d47f468622",
		Name:         "Oteclán",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenYouPlaneswalkToThisAndAtYourUpkeep("Oteclán — chaos ensues",
				func(g *game.Game, item *game.StackItem) error {
					g.ChaosEnsuesForEffect(item.Controller)
					return nil
				}),
			WheneverChaosEnsues("Oteclán — discover 3", DiscoverN(3)),
		},
	})
}
