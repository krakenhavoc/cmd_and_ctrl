package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Federation Probe — Artifact Creature — Construct {U}, 1/1 (Star Trek,
// #4973):
//
//	"Flying
//	 When this creature enters, choose one —
//	 • Target creature gets +1/+1 until end of turn.
//	 • Surveil 1."
//
// A modal enters trigger, Charming Prince's shape: a pump bullet and a
// surveil bullet.
//
// No simplification.
func init() {
	etb := WhenThisEnters("Federation Probe — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Target creature gets +1/+1 until end of turn.",
			TargetCreature("target creature"),
			BoostTheModesTarget(1, 1, "Federation Probe — +1/+1 until end of turn")),
		ModeDoing("Surveil 1.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Surveil{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:        "6fc7401f-66fe-4d17-be2f-afcf892be740",
		Name:            "Federation Probe",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
