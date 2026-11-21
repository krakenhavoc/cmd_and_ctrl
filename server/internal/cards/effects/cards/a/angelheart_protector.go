package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelheart Protector — Creature — Human Cleric {2}{W}, 3/2 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, target creature you control gains
//	 indestructible until end of turn."
//
// The grant is The Destined White Mage's target-keyword body
// (ebTargetCreatureGainsUntilEOT): chosen as the trigger goes on the
// stack, any creature its controller controls, the Protector included,
// and nothing if the target is no longer legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42aa599f-de2f-411a-8cf3-f5cfb40568b3",
		Name:         "Angelheart Protector",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Angelheart Protector — target creature you control gains indestructible until end of turn",
					effects.EbTargetCreatureGainsUntilEOT("indestructible", "Angelheart Protector — indestructible")),
				effects.TargetCreature("target creature you control", effects.YouControl()),
			),
		},
	})
}
