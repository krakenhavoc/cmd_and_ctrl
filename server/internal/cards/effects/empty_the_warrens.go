package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Empty the Warrens — Sorcery {3}{R}:
//
//	"Create two 1/1 red Goblin creature tokens.
//	 Storm (When you cast this spell, copy it for each spell cast
//	 before it this turn.)"
//
// The storm card with NO target clause, and the one that proves the
// keyword needs no per-card wiring for it. Its reminder text stops at
// "copy it for each spell cast before it this turn" — no "you may
// choose new targets", because there are none to choose — and
// `Storm()` still sets ChooseNewTargets unconditionally: the engine
// skips the CR 707.10c offer for a copy whose original named no
// target (itemHasChosenTarget, spell_copy.go), so this card gets no
// prompts and says nothing about it.
//
// Each copy is a separate resolution creating two tokens, so a storm
// count of four is five separate CR 701.7 creation instructions and
// ten Goblins. That is not the same as one instruction for ten: a
// doubler (Parallel Lives) applies to each, and the copies resolve one
// at a time with priority in between (#762, #1249 — the token
// creation path a copy of a token-making spell now takes).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a59c882-8bb8-49ba-862f-125020dd5bec",
		Name:         "Empty the Warrens",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: RedGoblinToken(), N: 2}.Apply(ctx)
		},
	})
}
