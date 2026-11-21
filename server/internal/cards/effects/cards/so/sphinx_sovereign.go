package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx Sovereign — Artifact Creature — Sphinx {4}{W}{U}{U}{B}, 6/6
// (Shards of Alara, #5588):
//
//	"Flying
//	 At the beginning of your end step, you gain 3 life if this creature
//	 is untapped. Otherwise, each opponent loses 3 life."
//
// The "if" is not an intervening one: it follows the effect, not the
// trigger condition (CR 603.4), so the ability always triggers and the
// Sphinx's tapped status is read as it resolves. A Sphinx that has left by
// then is judged as it last existed (CR 608.2h), Mana Vault's reading.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41907508-9ffc-4e01-969a-4f9c35cdf548",
		Name:            "Sphinx Sovereign",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStep("Sphinx Sovereign — gain 3 life if it is untapped, otherwise each opponent loses 3 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					info, ok := ctx.SourcePermanent()
					if ok && !info.Tapped {
						return effects.GainLife{Amount: 3}.Apply(ctx)
					}
					return effects.EachOpponentLosesLife{N: 3}.Apply(ctx)
				}),
		},
	})
}
