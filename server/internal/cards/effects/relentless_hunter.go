package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relentless Hunter — Creature — Human Warrior {1}{R}{G}, 3/3 (Oath of the Gatewatch, #5545):
//
//	"{1}{R}{G}: This creature gets +1/+1 and gains trample until end of
//	 turn."
//
// Flame-Chain Mauler's shape: the pump and the trample read the source
// at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6881e3ac-9e3d-49d2-b2ac-63feb6f0f2ff",
		Name:         "Relentless Hunter",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{G}: This creature gets +1/+1 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{R}{G}"),
			Effect:  thisCreatureUntilEOT("Relentless Hunter — +1/+1 and trample", 1, 1, "trample"),
		}},
	})
}
