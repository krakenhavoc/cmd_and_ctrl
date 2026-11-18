package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sepulcher Ghoul — Creature — Zombie {1}{B}, 2/1 (Adventures in the
// Forgotten Realms):
//
//	"Sacrifice another creature: This creature gets +2/+2 until end of
//	 turn. Activate only once each turn."
//
// "Another" is object identity (SacrificeAnotherN). The sacrifice is a
// cost, paid at the announce, so a creature that dies for it still pays
// even if the Ghoul is removed in response. "Once each turn" gates the
// offer (CR 602.5b).
//
// No simplification.
func init() {
	const label = "Sacrifice another creature: This creature gets +2/+2 until end of turn. Activate only once each turn."
	Register(Spec{
		OracleID:     "04936cc9-c915-4182-b84a-3491d26be548",
		Name:         "Sepulcher Ghoul",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerPump},
			Cost:      SacrificeAnotherN(1, "another creature", Creature()),
			Condition: OncePerTurnActivation(label),
			Effect:    thisGetsUntilEndOfTurn(2, 2, "Sepulcher Ghoul — +2/+2"),
		}},
	})
}
