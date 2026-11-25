package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Magic Mirror — Legendary Artifact {6}{U}{U}{U} (Throne of Eldraine,
// #5516):
//
//	"This spell costs {1} less to cast for each instant and sorcery card
//	 in your graveyard.
//	 You have no maximum hand size.
//	 At the beginning of your upkeep, put a knowledge counter on The Magic
//	 Mirror, then draw a card for each knowledge counter on The Magic
//	 Mirror."
//
// The discount is a self cost modifier (Cyan, Vengeful Samurai's shape), so
// it applies only while the Mirror is a spell. The upkeep trigger adds the
// counter, then draws as many cards as the Mirror has knowledge counters as
// it resolves (CR 608.2h: last-known when it has left).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "9acf2701-2ba1-46fa-ad4a-b130a58181d3",
		Name:          "The Magic Mirror",
		Completeness:  effects.CompletenessFull,
		NoMaxHandSize: true,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.SocB37InstantAndSorceryCardsInYourGraveyard,
				"This spell costs {1} less to cast for each instant and sorcery card in your graveyard."),
		},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("The Magic Mirror — put a knowledge counter on it, then draw a card for each knowledge counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.CounterOnThis{Kind: "knowledge", N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{N: effects.CountersOnThisAsItResolves("knowledge")(ctx)}.Apply(ctx)
				}),
		},
	})
}
