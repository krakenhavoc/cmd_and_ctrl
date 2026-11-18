package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a40de3b1-b24b-4da4-b550-9ac9838aecba",
		Name:         "Prosperous Partnership",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prosperous Partnership — create two 1/1 green and white Citizen creature tokens",
				Do(CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Tap three untapped creatures you control: Create a Treasure token.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  3,
				Filter: TargetPermanent("three untapped creatures you control", Creature()),
				Label:  "three untapped creatures you control",
			}},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TreasureToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
