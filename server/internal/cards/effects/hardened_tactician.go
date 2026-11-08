package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hardened Tactician — Creature — Human Warrior {1}{W}{B}, 2/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"{1}, Sacrifice a token: Draw a card."
//
// Fountainport's draw without the tap: the sacrifice is picked from the
// activator's own tokens of any kind (a Treasure, a Spirit, a mobilized
// Warrior), so the ability can be activated once per token.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca1dc6f0-3ec2-4f62-a08a-1395201bd93f",
		Name:         "Hardened Tactician",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice a token: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}"), b08SacrificeAToken()),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
