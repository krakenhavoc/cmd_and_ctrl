package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cranial Archive — Artifact {2} (Khans of Tarkir):
//
//	"{2}, Exile this artifact: Target player shuffles their graveyard into
//	 their library. Draw a card."
//
// Exiling is a cost, so the ability survives removal of the artifact. The
// target player shuffles; the Archive's controller draws.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "260f95c5-4058-4627-981f-98b9eafc2265",
		Name:         "Cranial Archive",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Exile this artifact: Target player shuffles their graveyard into their library. Draw a card.",
			Purpose: game.Purpose{Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.ExileThis()),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				who, ok := effects.FirstLegalPlayerTarget(ctx)
				if !ok {
					return nil
				}
				if err := effects.FinaleShuffleGraveyardIntoLibrary(ctx, who); err != nil {
					return err
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			},
		}},
	})
}
