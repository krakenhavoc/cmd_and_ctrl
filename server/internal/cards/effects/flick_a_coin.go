package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flick a Coin — Instant {2}{R}:
//
//	"Flick a Coin deals 1 damage to any target. You create a Treasure
//	 token.
//	 Draw a card."
//
// Three unconditional sentences in printed order. The Treasure and the
// card are not tied to the damage: a target that merely took the
// damage and died changes nothing here, while a spell whose only
// target became illegal is countered on resolution (CR 608.2b) and
// does none of it. (#3213, the Kuja deck.)
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "990583ef-5bf4-4e14-a8f1-61d32c6a75df",
		Name:         "Flick a Coin",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      ForTargets(DamageToTarget(0, 1)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToFirstTarget(1)(item, ctx); err != nil {
				return err
			}
			if err := (CreateToken{Controller: ctx.Controller(), Template: TreasureToken(), N: 1}).Apply(ctx); err != nil {
				return err
			}
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
