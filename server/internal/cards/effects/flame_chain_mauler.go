package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flame-Chain Mauler — Creature — Elemental Warrior {1}{R}, 2/2 (Lorwyn
// Eclipsed, #4458):
//
//	"{1}{R}: This creature gets +1/+0 and gains menace until end of turn.
//	 (It can't be blocked except by two or more creatures.)"
//
// Fangblade Brigand's firebreathing with menace in place of first strike
// (thisCreatureUntilEOT): both halves last until end of turn and do
// nothing if the Mauler has left the battlefield. Each activation adds
// another +1/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6eef6a7-0cf4-4ee8-90e0-727a283c3be9",
		Name:         "Flame-Chain Mauler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 and gains menace until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisCreatureUntilEOT("Flame-Chain Mauler — +1/+0 and menace", 1, 0, "menace"),
		}},
	})
}
