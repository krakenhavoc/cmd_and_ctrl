package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Overcome — Sorcery {3}{G}{G} (Ravnica: Clue Edition):
//
//	"Creatures you control get +2/+2 and gain trample until end of turn."
//
// The set of creatures is pinned as the spell resolves (CR 611.2c): one that
// enters afterwards gets nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15353e27-96ca-48c7-83d7-efd57c094643",
		Name:         "Overcome",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				match := effects.And(effects.Creature(), effects.YouControl())
				if err := (effects.BoostUntilEOT{Match: match, Power: 2, Toughness: 2, Label: "Overcome — +2/+2"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: match, Keywords: []string{"trample"}, Label: "Overcome — trample"}.Apply(ctx)
			}),
		),
	})
}
