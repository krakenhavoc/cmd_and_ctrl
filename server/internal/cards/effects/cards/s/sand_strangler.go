package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sand Strangler — {3}{R} Creature — Beast 3/3 (Hour of Devastation):
//
//	"When this creature enters, if you control a Desert or there is a Desert
//	 card in your graveyard, you may have this creature deal 3 damage to target
//	 creature."
//
// The "if" is an intervening condition (CR 603.4). The target is chosen as the
// trigger goes on the stack and the "may" is asked as it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c429716b-bd3a-4060-b8a3-82e3d6aecd88",
		Name:         "Sand Strangler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIfYou("you control a Desert or there is a Desert card in your graveyard", effects.OldB240YouHaveADesert,
					effects.Optional(
						effects.WhenThisEnters("Sand Strangler — it deals 3 damage to target creature", effects.SourceDealsDamageToEachLegalTarget(3)),
						"Sand Strangler — have it deal 3 damage to target creature?")),
				effects.TargetCreature("target creature")),
		},
	})
}
