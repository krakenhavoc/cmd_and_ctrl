package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lost in the Spirit World — Sorcery {2}{U} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"Return up to one target creature to its owner's hand. Create a 1/1
//	 colorless Spirit creature token with \"This token can't block or be
//	 blocked by non-Spirit creatures.\""
//
// The Spirit is Avatar Kuruk's printed row. The bounce is keyed so the
// token waits for any prompt the bounce asks; with no target chosen the
// spell still makes the Spirit, and with its one target gone it does not
// resolve at all (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a37adc18-ea11-4972-8d04-936cf3fe2613",
		Name:         "Lost in the Spirit World",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.BounceCardsToHandKeyedForEffect(legalTargetIDs(ctx), ctx.Tail(avB34ThenSpiritToken))
		},
	})
}
