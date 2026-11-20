package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weirding Shaman — Creature — Goblin Shaman {1}{B}, 2/1 (Morningtide,
// #5592):
//
//	"{3}{B}, Sacrifice a Goblin: Create two 1/1 black Goblin Rogue
//	 creature tokens."
//
// Pashalik Mons' activation in black: the Goblin is sacrificed as a cost
// (the Shaman itself may be it), and the tokens are Goblins, so they can
// pay for the next activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ed8e0c38-64b1-48a5-9623-935d59ffaacc",
		Name:         "Weirding Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{3}{B}, Sacrifice a Goblin: Create two 1/1 black Goblin Rogue creature tokens.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerMakesBlocker, Tokens: 2, CreatureTokens: 2},
			Cost:    Plus(ManaCost("{3}{B}"), b25SacrificeAGoblin()),
			Effect:  Do(CreateToken{Template: TokenCard("1/1 black Goblin Rogue"), N: 2}),
		}},
	})
}
