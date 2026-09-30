package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empty the Pits — Instant {X}{X}{B}{B}{B}{B}:
//
//	"Delve. Create X tapped 2/2 black Zombie creature tokens."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// {X}{X} is two slots, so X = 3 owes six generic, and delve may pay
// all six (ADR 0100 §1). The tokens are one creation instruction
// (CR 701.7b), so a token doubler doubles them. No simplification.
func init() {
	Register(Spec{
		OracleID:     "089d91cf-ed7a-4859-8967-cad975a5127e",
		Name:         "Empty the Pits",
		Completeness: CompletenessFull,
		Delve:        true,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateTokenAdvanced{
				Spec: Token(TokenCard("2/2 black Zombie")).EntersTapped(),
				N:    ctx.X(),
			}.Apply(ctx)
		},
	})
}
