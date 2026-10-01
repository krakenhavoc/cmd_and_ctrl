package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Trespass — Sorcery {8}{U}{U}{U}:
//
//	"Delve
//	 Take an extra turn after this one. Exile Temporal Trespass."
//
// ADR 0100 listed it as waiting on extra turns (#753) once sub-PR 1 had
// paid for its delve; ADR 0059's TakeExtraTurn has since landed, so the
// card is its two halves: `Delve: true` and Temporal Mastery's body.
// The spell exiles ITSELF as its last instruction, and #489's
// spellMovedItselfLocked stops the resolution frame from putting it in
// the graveyard afterwards. TakeExtraTurn queues and never fails.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c216b924-88ac-4853-9e95-0c345c09eeb6",
		Name:         "Temporal Trespass",
		Completeness: CompletenessFull,
		Delve:        true,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			_ = TakeExtraTurn{}.Apply(ctx)
			return ctx.Game.ExileCardForEffect(item.SourceCardID)
		},
	})
}
