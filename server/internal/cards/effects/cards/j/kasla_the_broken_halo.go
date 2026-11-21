package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kasla, the Broken Halo — Legendary Creature — Angel Ally {3}{U}{R}{W},
// 5/4 (March of the Machine Commander, #5647):
//
//	"Convoke
//	 Flying, vigilance, haste
//	 Whenever you cast another spell that has convoke, scry 2, then draw
//	 a card."
//
// Her own convoke is Spec.TapCost. The trigger asks whether the spell
// had convoke as it became cast (YouCastASpellThatHasConvoke), whatever
// gave it: printed, a permanent's grant or a next-spell promise.
// "Another" needs no check of its own, since Kasla is not on the
// battlefield while she is being cast. The draw waits for the scry's
// answer, so the card left on top is the card drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2d4ce3c-00b1-4a72-b652-4c129475080f",
		Name:            "Kasla, the Broken Halo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "haste"},
		TapCost:         effects.Convoke(),
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastASpellThatHasConvoke, "Kasla, the Broken Halo — scry 2, then draw a card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 1)}.Apply(ctx)
				}),
		},
	})
}
