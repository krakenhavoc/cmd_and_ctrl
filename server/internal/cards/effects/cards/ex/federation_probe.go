package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	etb := effects.WhenThisEnters("Federation Probe — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("Target creature gets +1/+1 until end of turn.",
			effects.TargetCreature("target creature"),
			effects.BoostTheModesTarget(1, 1, "Federation Probe — +1/+1 until end of turn")),
		effects.ModeDoing("Surveil 1.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Surveil{Player: item.Controller, N: 1}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:        "6fc7401f-66fe-4d17-be2f-afcf892be740",
		Name:            "Federation Probe",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
