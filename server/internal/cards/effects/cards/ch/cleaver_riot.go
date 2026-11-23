package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cleaver Riot — Sorcery {4}{R} (Magic 2013, #5568):
//
//	"Creatures you control gain double strike until end of turn. (They deal
//	 both first-strike and regular combat damage.)"
//
// A layer 6 grant (CR 613.1f) to the creatures you control as it
// resolves (CR 611.2c), so one that enters later this turn does not
// get it; it ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49f42896-90cd-4a05-bcca-12b19f780446",
		Name:         "Cleaver Riot",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"double strike"},
				Label: "Cleaver Riot — double strike"}.Apply(ctx)
		},
	})
}
