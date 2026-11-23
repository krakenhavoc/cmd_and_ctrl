package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prosperous Partnership — Enchantment {1}{R}{W} (Streets of New Capenna
// Commander):
//
//	"When this enchantment enters, create two 1/1 green and white Citizen
//	 creature tokens.
//	 Tap three untapped creatures you control: Create a Treasure token."
//
// Baylen's tap-others cost over creatures: it is not the {T} symbol (CR
// 302.6), so creatures that arrived this turn (the Citizens just made) pay it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a40de3b1-b24b-4da4-b550-9ac9838aecba",
		Name:         "Prosperous Partnership",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Prosperous Partnership — create two 1/1 green and white Citizen creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: effects.TargetPermanent("three untapped creatures you control", effects.Creature()),
				Label:  "three untapped creatures you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
