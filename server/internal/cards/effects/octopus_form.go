package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Octopus Form — Instant — Lesson {U}:
//
//	"Target creature you control gets +1/+1 and gains hexproof until end
//	 of turn. Untap it."
//
// Pump, keyword and untap on the one target, in the order printed.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "6fec368d-374d-4318-861b-ff478ba51a3c",
		Name:         "Octopus Form",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Octopus Form — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Octopus Form — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
