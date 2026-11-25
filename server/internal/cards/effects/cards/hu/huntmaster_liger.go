package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Huntmaster Liger — Creature — Cat {3}{W}, 3/4 (Ikoria: Lair of Behemoths,
// #5512):
//
//	"Mutate {2}{W} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Whenever this creature mutates, other creatures you control get +X/+X until
//	 end of turn, where X is the number of times this creature has mutated."
//
// Mutate is the engine's (ADR 0223). X is the merged permanent's count of
// mutations (S82OldB111TimesMutated), read as the trigger resolves, so a
// creature that has mutated twice gives +2/+2. The creatures it pumps are the
// ones you control as it resolves (CR 611.2c), never the Liger itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "38f5e6bf-18a3-4ff3-9726-a25a9715a2d9",
		Name:             "Huntmaster Liger",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{2}{W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Huntmaster Liger — other creatures you control get +X/+X until end of turn, where X is the number of times this creature has mutated",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					x := effects.S82OldB111XOfThisMutation(ctx)
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: x, Toughness: x,
						Label: "Huntmaster Liger — other creatures you control get +X/+X",
					}.Apply(ctx)
				}),
		},
	})
}
