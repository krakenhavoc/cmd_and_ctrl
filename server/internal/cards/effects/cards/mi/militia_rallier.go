package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Militia Rallier — Creature — Human Soldier {2}{W}, 3/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"This creature can't attack alone.
//	 Whenever this creature attacks, untap target creature."
//
// Pet Sehlat's attack limit (CR 506.5, 508.1c) and an attack trigger that
// untaps any creature, its own included; a target that left is not untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5377597-3454-4636-8563-7e6a5544e36a",
		Name:         "Militia Rallier",
		Completeness: effects.CompletenessFull,
		AttackLimits: []game.AttackLimit{effects.CantAttackAlone()},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Militia Rallier — untap target creature", effects.UntapFirstLegalTarget),
				effects.TargetCreature("target creature")),
		},
	})
}
