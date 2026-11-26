package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seasoned Marshal — Creature — Human Soldier {2}{W}{W}, 2/2 (Duel Decks:
// Elspeth vs. Tezzeret, #5578):
//
//	"Whenever this creature attacks, you may tap target creature."
//
// An attack trigger that takes its target as it goes on the stack (CR
// 603.3d) and asks its controller "you may" as it resolves (CR 603.5);
// declining taps nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0c7239bf-dc8a-4d79-867e-7a4225568c49",
		Name:         "Seasoned Marshal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WheneverThisAttacks("Seasoned Marshal — you may tap target creature", effects.TapTheFirstLegalTarget),
				effects.TargetCreature("target creature")),
				"Seasoned Marshal — tap the target creature?"),
		},
	})
}
