package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Localized Destruction — Sorcery {3}{W}{W}:
//
//	"You get {E} (an energy counter), then you may pay one or more {E}.
//	 If you do, each creature you control with power equal to the amount
//	 of {E} paid this way gains indestructible until end of turn.
//	 Destroy all creatures."
//
// ADR 0129 §3 (#1995, owner decision 3): "one or more" is the pay_amount
// prompt with a floor of 1. The creatures that gain indestructible are
// the ones with that power as the grant applies (CR 611.2c); then every
// creature is destroyed, in the same resolution. The bot is offered the
// power that saves the most of its creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "810e1371-e028-47e7-b97e-a627a35932e0",
		Name:         "Localized Destruction",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{Energy: 1,
			Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GetEnergy{N: 1}).Apply(ctx); err != nil {
				return err
			}
			return PayEnergyAmount{
				Min:      1,
				Question: "Localized Destruction — pay one or more {E}; your creatures with that power gain indestructible",
				Unit:     game.PayAmountPower,
				Goal:     powerThatSavesMostOfYours,
				Then: func(ctx *Context, paid int) error {
					if paid > 0 {
						if err := (GrantKeywordUntilEOT{
							Match:    And(Creature(), YouControl(), PowerGE(paid), PowerLE(paid)),
							Keywords: []string{"indestructible"},
							Label:    "Localized Destruction — indestructible until end of turn",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return DestroyAllMatching{Match: Creature()}.Apply(ctx)
				},
			}.Apply(ctx)
		},
	})
}
