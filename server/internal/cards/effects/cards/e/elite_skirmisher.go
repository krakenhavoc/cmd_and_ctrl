package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elite Skirmisher — Creature — Human Soldier {2}{W}, 3/1 (Foundations Jumpstart):
//
//	"Heroic — Whenever you cast a spell that targets this creature, you may
//	 tap target creature."
//
// The heroic trigger (CR 207.2c) is Brigone's; its own target is chosen as
// it goes on the stack and the "you may" is asked as it resolves (Choking
// Tethers' shape). A target gone by resolution is not tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66578806-9148-479d-80e1-6c3e3b67e452",
		Name:         "Elite Skirmisher",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.Heroic("Elite Skirmisher — you may tap target creature (heroic)", effects.TapTheFirstLegalTarget),
					"Elite Skirmisher — tap target creature?"),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
