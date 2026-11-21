package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sewer Crocodile — Creature — Crocodile {5}{U}, 4/6 (Streets of New Capenna):
//
//	"{3}{U}: This creature can't be blocked this turn. This ability costs
//	 {3} less to activate if there are five or more mana values among cards
//	 in your graveyard."
//
// The discount is the ability's OWN cost clause, read when the cost is
// determined (CR 601.2f) from the activating player's graveyard, and it
// spends against the generic part only ({U} always stays).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b01e9ff2-865a-401e-a839-568cd4a0451d",
		Name:         "Sewer Crocodile",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: This creature can't be blocked this turn. This ability costs {3} less to activate if there are five or more mana values among cards in your graveyard.",
			Cost:    effects.ManaCost("{3}{U}"),
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			CostModifiers: []game.CostModifier{
				effects.CostsLessEach(effects.SncCostLessThreeWithFive,
					"This ability costs {3} less to activate if there are five or more mana values among cards in your graveyard."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RestrictUntilEOT{
					Target:       item.SourceCardID,
					Restrictions: game.CantBeBlocked,
					Label:        "Sewer Crocodile — can't be blocked",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
