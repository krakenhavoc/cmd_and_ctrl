package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unyielding Gatekeeper — Creature — Elephant Cleric {1}{W} (Murders at
// Karlov Manor, #4469):
//
//	"Disguise {1}{W}
//	 When this creature is turned face up, exile another target nonland
//	 permanent. If you controlled it, return it to the battlefield tapped.
//	 Otherwise, its controller creates a 2/2 white and blue Detective
//	 creature token."
//
// Whose it was is read as the ability begins to resolve (the ruling), before
// the exile, so a permanent that changed hands since is judged by who
// controlled it then; the exile's keyed tail is told that player. A card that
// reached exile and was yours comes back tapped, under its owner's control
// (CR 400.7: an object returned from exile is new); a token you controlled
// ceased to exist and nothing returns. Anyone else's permanent costs its
// controller nothing but the permanent: they create the Detective, whether or
// not the exile happened.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "3c8fa799-7f30-4532-8317-f37a7ab787de",
		Name:             "Unyielding Gatekeeper",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{W}")},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisIsTurnedFaceUp("Unyielding Gatekeeper — exile another target nonland permanent; return it tapped if you controlled it, otherwise its controller creates a Detective",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						c, found := g.LookupCardForEffect(t.ID)
						if !found {
							return nil
						}
						tail := ctx.ItemTail(unyieldingGatekeeperAfter)
						tail.Carry = []uuid.UUID{c.Controller}
						return ExileTarget{Target: t.ID, ThenTail: tail}.Apply(ctx)
					}),
				Another(TargetPermanent("another target nonland permanent", Nonland()))),
		},
	})
}

// unyieldingGatekeeperAfter is the rest of the sentence, once the exile is
// done: Carry names who controlled the permanent as the ability resolved.
var unyieldingGatekeeperAfter = CardTail("tail/unyielding-gatekeeper-after-exile", func(ctx *Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	wasControlledBy := o.Tail.Carry[0]
	if wasControlledBy == ctx.Controller() {
		if len(o.IDs) == 0 {
			return nil
		}
		return ReturnFromExileTogether{Targets: o.IDs, Tapped: true}.Apply(ctx)
	}
	return CreateToken{Controller: wasControlledBy, Template: TokenCard("2/2 white and blue Detective"), N: 1}.Apply(ctx)
})
