package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Accursed Horde — Creature — Zombie {3}{B}, 3/3 (Game Night 2019, #5515):
//
//	"{1}{B}: Target attacking Zombie gains indestructible until end of turn."
//
// An attacking Zombie is a creature that is attacking now with the Zombie
// type, a changeling counted (CR 702.73a); the Horde may target itself. It is
// judged at announce and again as the ability resolves (CR 608.2b), so a
// Zombie removed from combat in response is not granted anything.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f3cba20-e7a4-47f4-8bce-90168ce9cc5c",
		Name:         "Accursed Horde",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: Target attacking Zombie gains indestructible until end of turn.",
			Cost:    effects.ManaCost("{1}{B}"),
			Targets: effects.TargetCreature("target attacking Zombie", effects.AttackingCreature(), effects.HasSubtype("Zombie")),
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")), game.AnswerProtect|game.AnswerSurvive|game.AnswerCombatGrant),
			Effect:  effects.PumpAndGrantFirstTarget(0, 0, "Accursed Horde — indestructible", "indestructible"),
		}},
	})
}
