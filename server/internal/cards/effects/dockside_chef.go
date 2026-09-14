package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "fed12a16-8920-403c-be63-0601a9d864b0",
		Name:         "Dockside Chef",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, Sacrifice an artifact or creature: Draw a card",
			Cost:  Plus(ManaCost("{1}{B}"), b30SacrificeAnArtifactOrCreature()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
