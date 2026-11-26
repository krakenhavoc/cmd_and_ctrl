package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shifting Loyalties — {5}{U} Sorcery (Fate Reforged):
//
//	"Exchange control of two target permanents that share a card type. (Artifact,
//	 creature, enchantment, land, and planeswalker are card types.)"
//
// Burglar's Plot's exchange over two clauses whose second pick is related to
// the first (SharesACardTypeWithTarget): the second is another permanent with a
// card type in common with the first, judged as the spell is cast (CR 601.2c)
// and again as it resolves (CR 608.2b), on the types after the layers. The
// exchange is all or nothing (CR 701.12a); neither half states a duration.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ee9a743-c678-4a76-b106-53f57a84848d",
		Name:         "Shifting Loyalties",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPermanent("target permanent"),
			effects.Distinct(effects.SharesACardTypeWithTarget(
				effects.TargetPermanent("another target permanent that shares a card type with it"), 0)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			a, okA := ctx.ClauseTarget(0)
			b, okB := ctx.ClauseTarget(1)
			if !okA || !okB || len(ctx.LegalTargets()) != 2 {
				return nil
			}
			return effects.ExchangeControl{A: a.ID, B: b.ID, Label: "Shifting Loyalties — exchange control"}.Apply(ctx)
		},
	})
}
