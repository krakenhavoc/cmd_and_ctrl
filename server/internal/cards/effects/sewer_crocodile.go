package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b01e9ff2-865a-401e-a839-568cd4a0451d",
		Name:         "Sewer Crocodile",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{3}{U}: This creature can't be blocked this turn. This ability costs {3} less to activate if there are five or more mana values among cards in your graveyard.",
			Cost:  ManaCost("{3}{U}"),
			CostModifiers: []game.CostModifier{
				CostsLessEach(sncCostLessThreeWithFive,
					"This ability costs {3} less to activate if there are five or more mana values among cards in your graveyard."),
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return RestrictUntilEOT{
					Target:       item.SourceCardID,
					Restrictions: game.CantBeBlocked,
					Label:        "Sewer Crocodile — can't be blocked",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
