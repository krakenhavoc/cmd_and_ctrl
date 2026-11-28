package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alaborn Cavalier — Creature — Human Knight {2}{W}{W}, 2/2 (Portal Second
// Age, S82 older sets batch 733, #5625):
//
//	"Whenever this creature attacks, you may tap target creature."
//
// The tap is chosen as the trigger goes on the stack and the "may" asked as
// it resolves (CR 603.5); a target that left in response is skipped
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8273146-f2e5-40ec-ba96-c0a4192caec8",
		Name:         "Alaborn Cavalier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(effects.On(game.EventAttack, effects.ThisAttacked,
					"Alaborn Cavalier — you may tap target creature", effects.TapTheFirstLegalTarget),
					"Alaborn Cavalier — tap target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
