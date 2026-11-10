package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Desculpting Blast — Instant {1}{U} (Edge of Eternities, #4461):
//
//	"Return target nonland permanent to its owner's hand. If it was
//	 attacking, create a 1/1 colorless Drone artifact creature token with
//	 flying and "This token can block only creatures with flying.""
//
// "Was attacking" is read as the spell begins to resolve, before the bounce
// takes the permanent out of combat (Prophesied End's reading). The token
// is the spell's controller's, and rides the bounce's tail so a pause the
// return asks for (a commander's CR 903.9 question) finishes first; it is
// made whether or not the permanent reached a hand, since the sentence is
// about the permanent, not about the return. The Drone is the Fae Court's
// Faerie shape: colourless, flying, with a block rule on the token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d80ee9c-55d4-4115-8a58-63e8460f9795",
		Name:         "Desculpting Blast",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					return nil
				}
				n := 0
				if c.AttackingTarget != uuid.Nil {
					n = 1
				}
				return BounceToHand{Target: t.ID, ThenTail: ctx.TailN(desculptingBlastDrone, n)}.Apply(ctx)
			}
			return nil
		},
	})
}

// desculptingBlastDrone is the bounce's "then": the Drone when the
// permanent was attacking (the tail's N). An on-disk key: never renamed,
// never reused.
var desculptingBlastDrone = CardTail("then/desculpting-blast-drone", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return CreateToken{Template: DroneToken(), N: 1}.Apply(ctx)
})
