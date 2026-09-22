package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Narset's Reversal — Instant {U}{U}:
//
//	"Copy target instant or sorcery spell, then return it to its
//	 owner's hand. You may choose new targets for the copy."
//
// Also NOT a counter, and for the same reason Reprieve's doc comment
// gives: the card never says "counter", so CR 701.6 does not apply.
// The original spell is copied (CR 707.10 — the copy resolves first,
// being on top) and then bounced to hand by the same
// ReturnSpellToHandForEffect Reprieve uses; a spell printed "can't be
// countered" is still both copied and returned, and no "whenever a
// spell is countered" payoff fires for either half.
//
// CopySpell runs first, matching the printed order ("copy … THEN
// return it") and CR 707.10a's mechanics: the copy is created on the
// stack above the original, which is still sitting there, untouched,
// for ReturnSpellToHand to find on the next line — the same paused-
// CR-903.9-leg shape counterSpellLocked's callers exercise, since
// both go through the identical stack-exit primitive
// (exitSpellFromStackLocked).
func init() {
	Register(Spec{
		OracleID:     "d55f6c70-321f-4fb4-bd33-0850ae1a7c36",
		Name:         "Narset's Reversal",
		Completeness: CompletenessFull,
		Targets:      instantOrSorcerySpell("target instant or sorcery spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// CR 608.2b: an illegal-target Reversal does not resolve at
			// all, and the engine has already countered it by game
			// rules before OnResolve runs — Reverberate's copy target
			// uses the same defensive shape.
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0].ID
			if err := (CopySpell{
				StackID:          target,
				Controller:       ctx.Controller(),
				ChooseNewTargets: true,
			}).Apply(ctx); err != nil {
				return err
			}
			return (ReturnSpellToHand{StackID: target}).Apply(ctx)
		},
	})
}
