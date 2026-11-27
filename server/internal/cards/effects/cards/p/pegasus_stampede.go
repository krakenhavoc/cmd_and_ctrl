package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pegasus Stampede — Sorcery {1}{W} (Exodus, S82 OLD batch 742, #5626):
//
//	"Buyback—Sacrifice a land. (You may sacrifice a land in addition to any
//	 other costs as you cast this spell. If you do, put this card into your
//	 hand as it resolves.)
//	 Create a 1/1 white Pegasus creature token with flying."
//
// Constant Mists' non-mana buyback (the land is sacrificed with the spell on
// the stack; the engine routes the card to hand as it resolves, CR 702.27a)
// over a one-token body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "16b4a48a-fec2-433c-948d-e0b16c961172",
		Name:         "Pegasus Stampede",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			effects.BuybackSacrifice("a land", effects.Land()),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 white Pegasus with flying"), N: 1}.Apply(ctx)
		},
	})
}
