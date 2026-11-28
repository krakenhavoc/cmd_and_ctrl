package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flanking Troops — Creature — Human Soldier {2}{W}{W}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Whenever this creature attacks, you may tap target creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47bbb630-2bf7-47a2-93bb-266e491803c3",
		Name:         "Flanking Troops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverThisAttacks("Flanking Troops — you may tap target creature", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature")),
				"Flanking Troops — tap the target creature?"),
		},
	})
}
