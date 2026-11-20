package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Daily Regimen — Enchantment — Aura {W} (Morningtide, #5592):
//
//	"Enchant creature
//	 {1}{W}: Put a +1/+1 counter on enchanted creature."
//
// Burden of Guilt's Aura-with-an-activation shape, with Level Up's "put a
// +1/+1 counter on enchanted creature" body (levelUpAuraEnters), which
// reads the host as the ability resolves and does nothing once the Aura
// is gone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a869ec0-c3ea-49a8-8586-d7e54e62a734",
		Name:         "Daily Regimen",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Activated: []ActivatedAbility{{
			Label:   "{1}{W}: Put a +1/+1 counter on enchanted creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{W}"),
			Effect:  levelUpAuraEnters,
		}},
	})
}
