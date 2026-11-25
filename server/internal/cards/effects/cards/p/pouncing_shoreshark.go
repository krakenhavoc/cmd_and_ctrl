package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pouncing Shoreshark — Creature — Shark Beast {4}{U}, 4/3 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {3}{U} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Flash
//	 Whenever this creature mutates, you may return target creature an opponent
//	 controls to its owner's hand."
//
// Mutate is the engine's (ADR 0223); the trigger fires on the merge whichever
// card ended on top (CR 702.140d). The target is chosen as the trigger goes on
// the stack (CR 603.3d) and the "you may" is asked as it resolves (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "38630f89-bc22-49b2-b97e-d31a5169decc",
		Name:             "Pouncing Shoreshark",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flash"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{3}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Pouncing Shoreshark — return target creature an opponent controls to its owner's hand", effects.BounceChosenTarget),
				effects.Harmful(effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))),
				"Return target creature an opponent controls to its owner's hand?"),
		},
	})
}
