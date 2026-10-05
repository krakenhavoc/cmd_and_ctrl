package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of Flame — Sorcery {R}:
//
//	"Add {R}{R}, then add {R} for each card named Rite of Flame in each
//	 graveyard."
//
// A sorcery, so it uses the stack and can be countered (CR 605.3a);
// the count is read as it resolves, across every graveyard, and the
// spell itself is still on the stack at that point, so it does not
// count itself — {R}{R} from the first copy, {R}{R}{R} from the second.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a2e53f9-8100-488f-8504-b59e9bd1cc29",
		Name:         "Rite of Flame",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			named := allGraveyardsCardIDs(ctx, func(c game.Card) bool { return c.Name == "Rite of Flame" })
			return AddMana{Produced: "{R}{R}" + strings.Repeat("{R}", len(named))}.Apply(ctx)
		},
	})
}
