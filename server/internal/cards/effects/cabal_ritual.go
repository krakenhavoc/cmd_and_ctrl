package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cabal Ritual — Instant {1}{B}:
//
//	"Add {B}{B}{B}.
//	 Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more
//	 cards in your graveyard."
//
// Dark Ritual with a threshold branch. Threshold is read as the spell
// resolves (SpellThreshold, the same condition Lightning Surge reads),
// so the card itself — on the stack, not in the graveyard — never
// counts toward its own seven.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b5bf1fa-6502-4790-b66b-f0f8504ebc7c",
		Name:         "Cabal Ritual",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if SpellThreshold()(ctx.Game, item) {
				return AddMana{Produced: "{B}{B}{B}{B}{B}"}.Apply(ctx)
			}
			return AddMana{Produced: "{B}{B}{B}"}.Apply(ctx)
		},
	})
}
