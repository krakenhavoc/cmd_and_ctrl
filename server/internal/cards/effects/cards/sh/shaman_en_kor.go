package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shaman en-Kor — Creature — Kor Cleric Shaman {1}{W}, 1/2:
//
//	"{0}: The next 1 damage that would be dealt to this creature this turn
//	 is dealt to target creature you control instead.
//	 {1}{W}: The next time a source of your choice would deal damage to
//	 target creature this turn, that damage is dealt to this creature
//	 instead."
//
// ADR 0108 §9 (#1905): the en-Kor row (enKorRow), and a "next time"
// redirection from a source chosen as the second ability resolves
// (CR 609.7a), from the target to the Shaman as it is now.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d9d5dbb-25ab-41a5-a277-3ee5f9ef579b",
		Name:         "Shaman en-Kor",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.EnKorRow(),
			effects.RedirectRow("{1}{W}: The next time a source of your choice would deal damage to target creature this turn, that damage is dealt to this creature instead.",
				effects.ManaCost("{1}{W}"), effects.TargetCreature("target creature"),
				effects.RedirectDamage{Choose: true, Protect: effects.ShieldTheTarget, Next: true, To: effects.RedirectToThis}),
		},
	})
}
