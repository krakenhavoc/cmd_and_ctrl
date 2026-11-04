package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Last Ronin's Technique — Instant {3}{W} (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Sneak {1}{W} (You may cast this spell for {1}{W} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Create three 1/1 white Ninja Turtle Spirit creature tokens. If this
//	 spell's sneak cost was paid, they enter tapped and attacking."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). "If this
// spell's sneak cost was paid" reads the resolving spell's own cast record
// (Context.PaidAltCost). Tokens put onto the battlefield attacking were
// never declared (CR 508.4), and their controller chooses what each
// attacks (CreateTokensAttackingYourChoice), the player, planeswalker or
// battle the returned creature was attacking offered first. Cast for its
// mana cost, the three tokens simply enter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "6a51bb21-0c01-4ad8-99d7-17685db90b69",
		Name:             "The Last Ronin's Technique",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{W}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			spirit := TokenCard("1/1 white Ninja Turtle Spirit")
			if !ctx.PaidAltCost(sneakAltCostKey) {
				return CreateToken{Template: spirit, N: 3}.Apply(ctx)
			}
			return CreateTokensAttackingYourChoice{
				Template: spirit,
				N:        3,
				Tapped:   true,
				Prefer:   ctx.ReturnedAttacking(),
				Label:    "The Last Ronin's Technique",
			}.Apply(ctx)
		},
	})
}
