package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lychguard — Artifact Creature — Necron {2}{B}, 2/3 (Warhammer 40,000
// Commander, S82 GAP batch 188, #5948):
//
//	"Guardian Protocols — {3}{B}, Sacrifice this creature: Return all
//	 legendary creature cards from your graveyard to your hand."
//
// The sacrifice is a cost, so Lychguard itself is in the graveyard when the
// ability resolves, and it is not legendary.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e82daa7b-d9fd-48c4-b5b2-f3b1f6ef1a24",
		Name:         "Lychguard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, Sacrifice this creature: Return all legendary creature cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnWholeGraveyardToHand(effects.NewContext(g, item), func(c game.Card) bool {
					return c.IsCreature() && c.IsLegendary()
				})
			},
		}},
	})
}
