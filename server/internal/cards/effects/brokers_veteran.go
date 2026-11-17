package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Broker's Veteran — Creature — Human Soldier {1}{U}, 2/1 (Streets of
// New Capenna):
//
//	"When this creature dies, put a shield counter on target creature
//	 you control."
//
// The target is chosen as the trigger goes on the stack; with it gone
// by resolution nothing happens (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "feee1926-0b63-404f-9b01-df965de26b5f",
		Name:         "Brokers Veteran",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Brokers Veteran — put a shield counter on target creature you control",
				Do(Instruction(snc6ShieldOnFirstTarget))),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
