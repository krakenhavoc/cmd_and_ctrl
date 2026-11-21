package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gemrazer — Creature — Beast {3}{G}, 4/4 (Ikoria: Lair of Behemoths;
// reprinted in Fallout, #5948):
//
//	"Mutate {1}{G}{G} (If you cast this spell for its mutate cost, put
//	 it over or under target non-Human creature you own. They mutate
//	 into the creature on top plus all abilities from under it.)
//	 Reach, trample
//	 Whenever this creature mutates, destroy target artifact or
//	 enchantment an opponent controls."
//
// Everquill Phoenix's mutate (Mutate, ADR 0223) and its trigger on the
// merge (ThisCreatureMutates), which fires whichever card ended on top and
// not when the Gemrazer is cast for its mana cost. The target is chosen as
// the trigger goes on the stack (CR 603.3d), so a merge with no artifact or
// enchantment an opponent controls puts nothing on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "3dfb0c0a-b68f-43b9-8475-28d0192fc4ed",
		Name:             "Gemrazer",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"reach", "trample"},
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{1}{G}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Gemrazer — destroy target artifact or enchantment an opponent controls", effects.DestroyChosenPermanent),
				effects.TargetPermanent("target artifact or enchantment an opponent controls", effects.Or(effects.Artifact(), effects.Enchantment()), effects.OpponentControls())),
		},
	})
}
