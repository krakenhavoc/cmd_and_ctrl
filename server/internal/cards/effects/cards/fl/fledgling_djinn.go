package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fledgling Djinn — Creature — Djinn {1}{B}, 2/2 (Weatherlight, #5629):
//
//	"Flying
//	 At the beginning of your upkeep, this creature deals 1 damage to you."
//
// Serendib Efreet's two lines. The damage is dealt by the Djinn even
// when it has left the battlefield in response, as it last existed
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2c73ef77-ae58-401f-8747-538c4cd075d0",
		Name:            "Fledgling Djinn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Fledgling Djinn — this creature deals 1 damage to you",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DealDamage{Source: item.SourceCardID, Target: item.Controller, Amount: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
