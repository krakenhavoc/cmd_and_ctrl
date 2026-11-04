package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emergency Medical Hologram — {W} Artifact Creature — Illusion Doctor
// 0/1 (Star Trek, #4973):
//
//	"When this creature enters, put a +1/+1 counter on target creature.
//	 You gain 1 life."
//
// Ironshell Beetle's trigger with the life gain after it. The target is
// any creature, the Hologram included. If the only target is illegal as
// the trigger resolves, the whole ability does nothing and no life is
// gained (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8170dcbb-9879-414c-819d-ea2658ff770c",
		Name:         "Emergency Medical Hologram",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Emergency Medical Hologram — put a +1/+1 counter on target creature, then gain 1 life",
				Do(Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }),
					GainLife{Amount: 1})),
				TargetCreature("target creature")),
		},
	})
}
