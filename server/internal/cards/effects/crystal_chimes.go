package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Chimes — Artifact {3} (Urza's Saga):
//
//	"{3}, {T}, Sacrifice this artifact: Return all enchantment cards from
//	 your graveyard to your hand."
//
// Praetor's Counsel's return, narrowed to enchantment cards. The cards
// are read as the ability resolves, so an enchantment put into your
// graveyard in response comes back too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbff7452-1da4-4cd0-8247-a625e4b4da2b",
		Name:         "Crystal Chimes",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: Return all enchantment cards from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return returnWholeGraveyardToHand(NewContext(g, item), game.Card.IsEnchantment)
			},
		}},
	})
}
