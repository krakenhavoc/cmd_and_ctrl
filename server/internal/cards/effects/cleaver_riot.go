package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "49f42896-90cd-4a05-bcca-12b19f780446",
		Name:         "Cleaver Riot",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"double strike"},
				Label: "Cleaver Riot — double strike"}.Apply(ctx)
		},
	})
}
