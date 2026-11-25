package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Territorial Allosaurus — Creature — Dinosaur {2}{G}{G}, 5/5 (Dominaria):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this spell.)
//	 When this creature enters, if it was kicked, it fights another target
//	 creature."
//
// A kicked enters trigger (an intervening "if", CR 603.4, read off the cast
// record) whose target is another creature, chosen as the trigger goes on the
// stack; the fight is Aggressive Biomancy's, each creature dealing damage equal
// to its power to the other (CR 701.14). If the Allosaurus has gone by then, or
// the target has, nothing fights (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "7fe39c74-cb7c-4eee-a657-8a99a4ab383a",
		Name:          "Territorial Allosaurus",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{2}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Territorial Allosaurus — kicked, it fights another target creature", effects.ItFightsTheFirstTarget),
				effects.Another(effects.TargetCreature("another target creature"))),
		},
	})
}
