package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "158aaeab-1646-4160-9615-bc8747620a03",
		Name:          "Choking Miasma",
		Purpose:       game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{G}")},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return ChoosePermanents{
					Question:   "Choking Miasma — put a +1/+1 counter on a creature you control",
					Candidates: humiliateCreatures,
					Keyed:      chokingMiasmaChosen,
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				return BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2, Label: "Choking Miasma — -2/-2 until end of turn"}.Apply(ctx)
			}),
		),
	})
}

var chokingMiasmaChosen = RegisterPermanentsPick("prompt-run/choking-miasma-counter", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})
