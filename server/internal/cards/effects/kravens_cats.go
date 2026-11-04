package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kraven's Cats — Creature — Cat Villain {1}{G}, 2/2 (Marvel's Spider-Man,
// #4460):
//
//	"{2}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn."
//
// The limit counts activations (CR 602.1b), as Barbarian Bully's does, and
// the pump is the shared self pump.
//
// No simplification.
func init() {
	const label = "{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "976db753-9887-448a-8e67-ec7937b6505c",
		Name:         "Kraven's Cats",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      ManaCost("{2}{G}"),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Kraven's Cats — +2/+2 until end of turn"),
		}},
	})
}
