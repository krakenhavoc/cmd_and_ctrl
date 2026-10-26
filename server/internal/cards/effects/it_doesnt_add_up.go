package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// It Doesn't Add Up — Instant {3}{B}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield. Suspect it. (It has menace and can't block.)"
//
// Zombify's return, then the suspect on the permanent that came back.
// "It" is the NEW object (CR 400.7), so the suspect is aimed at the id
// the return reports, not at the card's old one.
//
// The suspect is the entry's tail (ReturnFromGraveyardToBattlefieldThenForEffect, ADR 0154),
// so it runs inline when the creature simply enters and from the
// player's answer when the entry first asks something — a returned
// Clone's copy pick, a returned shockland's payment — and it lands on
// the permanent that is on the battlefield by then.
func init() {
	Register(Spec{
		OracleID:     "dcfc8c59-d032-414d-9da6-125ebcb1366c",
		Name:         "It Doesn't Add Up",
		Completeness: CompletenessFull,
		Targets:      targetCreatureInYourGraveyard(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			// CR 608.2b: a creature that left the graveyard in response is
			// told to the tail as uuid.Nil, and nothing is suspected.
			return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(target, ctx.Controller(), false, ctx.Tail(thenSuspectEach))
		},
	})
}
