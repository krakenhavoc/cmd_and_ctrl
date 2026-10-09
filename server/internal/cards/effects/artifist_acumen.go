package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artifist Acumen — Sorcery {R}:
//
//	"Creatures you control gain first strike until end of turn.
//	 Draw a card."
//
// The set of creatures is fixed as the spell resolves (CR 611.2c), so a
// creature that enters later this turn does not get first strike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "129b41cf-8327-42f0-b186-1111b42c08ab",
		Name:         "Artifist Acumen",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := (GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"first strike"},
				Label:    "Artifist Acumen — first strike",
			}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
