package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reprieve — Instant {1}{W}:
//
//	"Return target spell to its owner's hand.
//	 Draw a card."
//
// NOT a counter. The card never prints the word, so CR 701.6 does not
// apply to it: a spell printed "can't be countered" is still bounced,
// and nothing watching "whenever a spell is countered" fires. That is
// the whole reason ReturnSpellToHand exists beside CounterTarget
// rather than as a destination on it (#1230) — game.counterSpellLocked
// already accepted a destination zone, but reusing it here would have
// carried its Countered flag and its "can't be countered" gate along
// for free, both wrong for this card.
//
// game.ReturnSpellToHandForEffect shares counterSpellLocked's
// stack-exit body instead: the same S29 flashback override and the
// same CR 614 / CR 903.9 window through routeCardToZoneLocked, with
// Countered set to false. A commander answered by Reprieve still gets
// the CR 903.9 offer before it lands in a hand, exactly as a countered
// one would before landing in a graveyard.
func init() {
	Register(Spec{
		OracleID:     "f0449cbd-855c-4c20-a1f5-f76a395d8d39",
		Name:         "Reprieve",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			// CR 608.2b: an illegal-target Reprieve does not resolve at
			// all, and the engine has already countered it by game
			// rules before OnResolve runs. This guard is the same
			// defensive shape Reverberate's copy target uses.
			if len(item.Targets) == 0 {
				return nil
			}
			if err := (ReturnSpellToHand{StackID: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
