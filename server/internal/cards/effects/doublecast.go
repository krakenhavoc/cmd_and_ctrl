package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doublecast — Sorcery {R}{R}:
//
//	"When you next cast an instant or sorcery spell this turn, copy
//	 that spell. You may choose new targets for the copy."
//
// The card #663 was opened for, and the whole reason ADR 0026's
// "a delayed trigger has no event watch" needed amending: there is no
// step to wait for. It waits for a CAST, fires on the first one that
// matches (CR 603.7b), and ends with the turn if none ever comes
// (CR 514.2).
//
// Three things it is NOT, each of which was a tempting shortcut:
//
//   - It is not Reverberate. Reverberate targets a spell already on
//     the stack; Doublecast has no target at all and is cast BEFORE
//     the spell it copies exists. A sorcery, so it is also cast a
//     turn-phase earlier than the instant it usually doubles.
//   - It is not a replacement on casting. The copy is put on the
//     stack by a triggered ability that goes on the stack above the
//     spell, so the table gets a response window and the copy
//     resolves FIRST.
//   - It does not copy itself. The trigger is created while Doublecast
//     resolves, and the EventCast of the spell that made it was
//     emitted before that resolution began — the ordering does the
//     work, so nothing has to remember the exclusion.
//
// "You may choose new targets for the copy" is CR 707.10c and rides
// CopySpell.ChooseNewTargets, the same clause Reverberate prints.
func init() {
	Register(Spec{
		OracleID:     "946675f5-8998-4f1b-934b-85ffe6e2f002",
		Name:         "Doublecast",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return WhenYouNextCast(
				"Doublecast — copy that spell",
				game.CastFilter{Types: []string{"Instant", "Sorcery"}},
				copyTheSpellBody,
			).Apply(ctx)
		},
	})
}
