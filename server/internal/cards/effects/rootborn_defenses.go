package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rootborn Defenses — Instant {2}{W}:
//
//	"Populate. Creatures you control gain indestructible until end of
//	 turn. (To populate, create a token that's a copy of a creature
//	 token you control.)"
//
// The order is the rules content: populate first, so the new token is
// among "creatures you control" when the grant is made (CR 611.2c
// fixes the set at resolution) and is protected too. The grant is the
// populate's continuation because the populate can open a prompt, and
// a line after it would run before the player has chosen.
//
// With no creature token to copy the populate does nothing (CR
// 701.36b) and the indestructible grant still happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04046ae4-5c51-434b-930c-f3b1d348bf4b",
		Name:         "Rootborn Defenses",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Populate{ThenTail: ctx.ItemTail(rootbornDefensesGrant)}.Apply(ctx)
		},
	})
}

// rootbornDefensesGrant is the indestructible grant that follows the
// populate, keyed (#4253): the populate can open a prompt, and the grant
// waits for it. An on-disk key: never renamed, never reused.
var rootbornDefensesGrant = CardTail("rootborn-defenses/grant-indestructible", func(ctx *Context, _ game.TailOutcome) error {
	return GrantKeywordUntilEOT{
		Match:    And(Creature(), YouControl()),
		Keywords: []string{"indestructible"},
		Label:    "Rootborn Defenses — creatures you control gain indestructible",
	}.Apply(ctx)
})
