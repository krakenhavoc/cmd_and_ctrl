package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pegasus Refuge — Enchantment {3}{W} (Tempest, S82 OLD b752, #5628):
//
//	"{2}, Discard a card: Create a 1/1 white Pegasus creature token with
//	 flying."
//
// Chivalric Alliance's cost, {2} plus discarding a card (paid as the ability is
// activated, CR 602.2b), then one token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07462467-e4a3-409e-bcef-9cc92ca4c299",
		Name:         "Pegasus Refuge",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Discard a card: Create a 1/1 white Pegasus creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{
					Controller: item.Controller,
					Template:   effects.TokenCard("1/1 white Pegasus with flying"),
					N:          1,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
