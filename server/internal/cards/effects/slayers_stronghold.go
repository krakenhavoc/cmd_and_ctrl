package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slayers' Stronghold — Land (#3656):
//
//	"{T}: Add {C}.
//	 {R}{W}, {T}: Target creature gets +2/+0 and gains vigilance and
//	 haste until end of turn."
//
// Reckless Stormseeker's boost-and-keywords pair on an activated
// ability; a target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2de7367b-a5a4-43f4-8f8f-b931ea28150d",
		Name:         "Slayers' Stronghold",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}{W}, {T}: Target creature gets +2/+0 and gains vigilance and haste until end of turn.",
			Cost:    Plus(ManaCost("{R}{W}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				target, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				if err := (BoostUntilEOT{Target: target, Power: 2, Label: "Slayers' Stronghold — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Target: target, Keywords: []string{"vigilance", "haste"}, Label: "Slayers' Stronghold — vigilance and haste"}.Apply(ctx)
			},
		}},
	})
}
