package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feldon's Cane — Artifact {1} (EDHREC rank 7538):
//
//	"{T}, Exile this artifact: Shuffle your graveyard into your
//	 library."
//
// The one-shot decking insurance. The cost is {T} plus #1404's
// ExileThis() paid from the BATTLEFIELD, so the Cane is gone at
// announce and cannot be shuffled in with the graveyard it resets —
// it is in exile, not the graveyard, by the time the ability resolves.
// The effect is the same "every card in your graveyard, then shuffle"
// Finale of Revelation's ten-or-more branch uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9b884dfd-59f4-45c0-bf1e-6ad9f5b58895",
		Name:         "Feldon's Cane",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}, Exile this artifact: Shuffle your graveyard into your library.",
			Cost:  Plus(TapCost(), ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return finaleShuffleGraveyardIntoLibrary(NewContext(g, item), item.Controller)
			},
		}},
	})
}
