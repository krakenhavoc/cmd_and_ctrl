package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aftermath Analyst — Creature — Elf Detective {1}{G}, 1/3 (EDHREC
// rank 631):
//
//	"When this creature enters, mill three cards.
//	 {3}{G}, Sacrifice this creature: Return all land cards from your
//	 graveyard to the battlefield tapped."
//
// Self-mill on the way in, Splendid Reclamation on the way out. The
// ETB is an ordinary trigger; the cash-in is a CR 602 activated
// ability whose cost is mana plus sacrificing the source (Commander's
// Sphere's shape with a mana component), and its effect takes every
// land card in the controller's graveyard, read once before anything
// moves, and returns them (#1867) as one entry, tapped on the entry
// event: each land sees the others enter (CR 603.6a). Landfall
// triggers fire once per land, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "374e54d8-8e73-4268-9dde-c28c77bbbf32",
		Name:         "Aftermath Analyst",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aftermath Analyst — mill three cards", Do(MillCards{N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}{G}, Sacrifice this creature: Return all land cards from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}{G}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b10ReturnAllLandCardsFromGraveyardTapped(NewContext(g, item), item.Controller)
			},
		}},
	})
}
