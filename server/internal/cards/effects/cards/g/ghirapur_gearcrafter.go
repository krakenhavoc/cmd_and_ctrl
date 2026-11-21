package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ghirapur Gearcrafter — Creature — Human Artificer {2}{R}, 2/1 (Magic Origins, #5549):
//
//	"When this creature enters, create a 1/1 colorless Thopter artifact creature token with flying. (A creature with flying can't be blocked except by creatures with flying or reach.)"
//
// Nimble Thopterist's entry token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "710dc9b1-7d32-4330-95a8-3e91a668202a",
		Name:         "Ghirapur Gearcrafter",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Ghirapur Gearcrafter — create a 1/1 colorless Thopter artifact creature token with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 1})),
		},
	})
}
