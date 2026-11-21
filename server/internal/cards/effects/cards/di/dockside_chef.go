package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dockside Chef — Enchantment Creature — Human Citizen {B}, 1/2
// (EDHREC rank 3243):
//
//	"{1}{B}, Sacrifice an artifact or creature: Draw a card."
//
// The one-mana sac outlet that cantrips. The cost is the shared
// sacrifice-other shape over any artifact or creature the activator
// controls — the Chef itself qualifies, being a creature — paid at
// announce, so the sacrificed permanent's dies-triggers land above
// the draw and resolve first (Blood Artist drains before the card
// is drawn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fed12a16-8920-403c-be63-0601a9d864b0",
		Name:         "Dockside Chef",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, Sacrifice an artifact or creature: Draw a card",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.B30SacrificeAnArtifactOrCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
