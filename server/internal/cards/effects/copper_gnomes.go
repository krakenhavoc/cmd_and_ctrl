package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Copper Gnomes — Artifact Creature — Gnome {2}, 1/1 (Urza's Saga):
//
//	"{4}, Sacrifice this creature: You may put an artifact card from your
//	 hand onto the battlefield."
//
// Master Transmuter's put with Copper Gnomes's own cost. The pick is made
// as the ability resolves, from artifact cards in hand only, and
// declining puts nothing. With no {T} in the cost, the Gnomes can be
// sacrificed the turn they arrive.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fd6d9af-4d53-4e10-bc23-672454803099",
		Name:         "Copper Gnomes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, Sacrifice this creature: You may put an artifact card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), SacrificeThis()),
			Effect: Do(PutFromHandOntoBattlefield{
				Match:    Artifact(),
				Optional: true,
				Label:    "Copper Gnomes — you may put an artifact card from your hand onto the battlefield",
			}),
		}},
	})
}
