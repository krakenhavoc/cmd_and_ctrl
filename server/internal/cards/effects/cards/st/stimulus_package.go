package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ce76d17e-d536-46f0-98e2-3e1b1b4f61a4",
		Name:         "Stimulus Package",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Stimulus Package — create two Treasure tokens",
				effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a Treasure: Create a 1/1 green and white Citizen creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("a Treasure", effects.IsTreasure)},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("1/1 green and white Citizen"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
