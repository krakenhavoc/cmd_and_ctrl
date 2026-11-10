package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wanderwine Farewell — Kindred Sorcery — Merfolk {5}{U}{U} (Lorwyn
// Eclipsed, #4458):
//
//	"Convoke
//	 Return one or two target nonland permanents to their owners' hands.
//	 Then if you control a Merfolk, create a 1/1 white and blue Merfolk
//	 creature token for each permanent returned to its owner's hand this
//	 way."
//
// The returns are one simultaneous batch with a tail (CR 400.7): the tail
// is told which cards actually reached a hand (a commander that took the
// command zone, or a move a replacement redirected, is not one), counts
// them, and only then looks at the board for a Merfolk, so a Merfolk that
// was itself returned does not count and the tokens themselves are made
// after the check. Convoke is the shared tap cost (CR 702.51).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7e1da4a1-5922-4a2e-8bbc-25fbcbf20935",
		Name:         "Wanderwine Farewell",
		Completeness: CompletenessFull,
		TapCost:      Convoke(),
		Targets:      TargetPermanent("one or two target nonland permanents", Nonland()).WithCount(1, 2),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.BounceCardsToHandKeyedForEffect(ids, ctx.ItemTail(wanderwineFarewellTokens))
		},
	})
}

// wanderwineFarewellTokens is "then if you control a Merfolk, create a
// 1/1 white and blue Merfolk creature token for each permanent returned
// this way". An on-disk key.
var wanderwineFarewellTokens = CardTail("wanderwine-farewell/merfolk-tokens", func(ctx *Context, o game.TailOutcome) error {
	n := len(o.IDs)
	if n == 0 {
		return nil
	}
	merfolk := false
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtypeWord("Merfolk") {
			merfolk = true
			break
		}
	}
	if !merfolk {
		return nil
	}
	return CreateToken{Template: TokenCard("1/1 white and blue Merfolk"), N: n}.Apply(ctx)
})
