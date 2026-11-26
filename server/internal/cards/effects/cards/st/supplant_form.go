package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Supplant Form — {4}{U}{U} Instant (Fate Reforged):
//
//	"Return target creature to its owner's hand. You create a token that's a
//	 copy of that creature."
//
// The copy is made as the bounce finishes, from the creature's last-known
// copiable values (CR 608.2h, 707.2), so a creature that was itself a copy is
// copied as that copy. The token is yours whoever owned the creature, and is
// created whether or not the return happened (a replacement may have sent the
// card elsewhere). A target that is no longer legal makes the spell do nothing
// (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e49f689-996a-477e-a327-32b69fbc0d85",
		Name:         "Supplant Form",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id := effects.FirstLegalBattlefieldTarget(ctx)
			if id == uuid.Nil {
				return nil
			}
			// The copy is made from the bounce's continuation, so it waits for a
			// CR 903.9 question the return asks; it is not gated on the creature
			// having reached a hand ("You create a token" is its own sentence).
			then := ctx.ItemTail(supplantFormCopy)
			then.Carry = []uuid.UUID{id}
			return effects.BounceToHand{Target: id, ThenTail: then}.Apply(ctx)
		},
	})
}

// supplantFormCopy is "You create a token that's a copy of that creature": the
// creature's last-known copiable values (CR 608.2h, 707.2), read through
// TokenCopySourceForEffect, which finds the permanent that left. An on-disk
// key: never renamed, never reused.
var supplantFormCopy = effects.CardTail("then/s82-old-b340-supplant-form-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	tmpl, ok := effects.TokenCopyTemplate(ctx.Game, o.Tail.Carry[0])
	if !ok {
		return nil
	}
	return effects.CreateTokenCopies(ctx, ctx.Controller(), tmpl, 1)
})
