package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stimulus Package — Enchantment {2}{R}{G} (Streets of New Capenna):
//
//	"When this enchantment enters, create two Treasure tokens.
//	 Sacrifice a Treasure: Create a 1/1 green and white Citizen creature
//	 token."
//
// The sacrifice is Jolene's "Sacrifice a Treasure" cost, with no mana; it
// is paid on activation, so the ability resolves even if the Package has
// left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce76d17e-d536-46f0-98e2-3e1b1b4f61a4",
		Name:         "Stimulus Package",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Stimulus Package — create two Treasure tokens",
				Do(CreateToken{Template: TreasureToken(), N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Treasure: Create a 1/1 green and white Citizen creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Cost:    game.AbilityCost{SacrificeOther: sacrificeSpec("a Treasure", isTreasure)},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
