package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunshower Druid — Creature — Frog Druid {G}, 0/2 (Bloomburrow, #4467):
//
//	"When this creature enters, put a +1/+1 counter on target creature
//	 and you gain 1 life."
//
// Vitu-Ghazi Inspector's counter and life without the evidence: the
// counter first, then the life. The target is any creature, the Druid
// itself included. A target that left in response removes the trigger
// (CR 608.2b), and then no life is gained either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7a70faf5-cc04-402b-92e3-67746788f510",
		Name:         "Sunshower Druid",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sunshower Druid — put a +1/+1 counter on target creature and gain 1 life",
				Do(Instruction(func(ctx *Context) error { return PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }), GainLife{Amount: 1})),
				TargetCreature("target creature")),
		},
	})
}
