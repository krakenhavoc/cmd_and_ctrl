package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarkhan Vol — Legendary Planeswalker — Sarkhan {2}{R}{G}, loyalty 4:
//
//	"+1: Creatures you control get +1/+1 and gain haste until end of turn.
//	 −2: Gain control of target creature until end of turn. Untap that
//	 creature. It gains haste until end of turn.
//	 −6: Create five 4/4 red Dragon creature tokens with flying."
//
// +1 is decided as it resolves (CR 611.2c): the creatures you control then.
// −2 is control, the untap and haste in that order, each for this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d902074a-2a2b-42b7-a1e5-750862c64fbe",
		Name:            "Sarkhan Vol",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Activated: []effects.ActivatedAbility{
			{
				Label: "+1: Creatures you control get +1/+1 and gain haste until end of turn.",
				Cost:  effects.LoyaltyCost(1),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					yours := effects.And(effects.Creature(), effects.YouControl())
					if err := (effects.BoostUntilEOT{Match: yours, Power: 1, Toughness: 1, Label: "Sarkhan Vol — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"haste"}, Label: "Sarkhan Vol — haste"}.Apply(ctx)
				})),
			},
			{
				Label:   "−2: Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.ThreatenAndGrant(ctx, id, "Sarkhan Vol", "haste")
				})),
			},
			{
				Label:   "−6: Create five 4/4 red Dragon creature tokens with flying.",
				Cost:    effects.LoyaltyCost(-6),
				Purpose: game.Purpose{Tokens: 5},
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/4 red Dragon with flying"), N: 5}),
			},
		},
	})
}
