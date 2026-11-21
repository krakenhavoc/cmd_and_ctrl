package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choking Miasma — Sorcery {1}{B}{B} (Dominaria United):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	 If this spell was kicked, put a +1/+1 counter on a creature you
//	 control.
//	 All creatures get -2/-2 until end of turn."
//
// Printed order is resolution order (CR 608.2c): the kicked counter
// goes on first, so the creature it lands on is already bigger when the
// -2/-2 arrives. The counter is not targeted: the creature is chosen as
// the spell resolves, and with no creature to choose the clause does
// nothing. The -2/-2 locks the set of creatures as it begins
// (CR 611.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:      "158aaeab-1646-4160-9615-bc8747620a03",
		Name:          "Choking Miasma",
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{G}")},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.ChoosePermanents{
					Question:   "Choking Miasma — put a +1/+1 counter on a creature you control",
					Candidates: effects.HumiliateCreatures,
					Keyed:      chokingMiasmaChosen,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.BoostUntilEOT{Match: effects.Creature(), Power: -2, Toughness: -2, Label: "Choking Miasma — -2/-2 until end of turn"}.Apply(ctx)
			}),
		),
	})
}

var chokingMiasmaChosen = effects.RegisterPermanentsPick("prompt-run/choking-miasma-counter", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
