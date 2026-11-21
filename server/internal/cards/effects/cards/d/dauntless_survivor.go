package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dauntless Survivor — Creature — Human Warrior {1}{G}, 1/1 (Zendikar
// Rising, #5508):
//
//	"When this creature enters, put a +1/+1 counter on target creature."
//
// Jeong Jeong's Deserters' trigger: any creature, the Survivor itself
// included, chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8371dee1-5f4a-410a-8b7c-e4110cdbe9bd",
		Name:         "Dauntless Survivor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dauntless Survivor — put a +1/+1 counter on target creature", effects.RfCreatureFCounterOnTargetCreature),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
