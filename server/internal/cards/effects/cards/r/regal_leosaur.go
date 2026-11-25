package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Regal Leosaur — Creature — Dinosaur Cat {R}{W}, 2/2 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {1}{R/W}{R/W} (If you cast this spell for its mutate cost, put it
//	 over or under target non-Human creature you own. They mutate into the
//	 creature on top plus all abilities from under it.)
//	 Whenever this creature mutates, other creatures you control get +2/+1 until
//	 end of turn."
//
// Mutate is the engine's (ADR 0223). The pump goes to the creatures you
// control as the trigger resolves (CR 611.2c), never the merged creature
// itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "fc445b1d-229d-4cf3-8203-8872d8957a15",
		Name:             "Regal Leosaur",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{1}{R/W}{R/W}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates, "Regal Leosaur — other creatures you control get +2/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), Power: 2, Toughness: 1,
						Label: "Regal Leosaur — other creatures you control get +2/+1",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
