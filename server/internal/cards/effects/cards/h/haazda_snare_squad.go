package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haazda Snare Squad — Creature — Human Soldier {2}{W}, 1/4 (Dragon's Maze,
// S82 OLD b405, #5564):
//
//	"Whenever this creature attacks, you may pay {W}. If you do, tap target
//	 creature an opponent controls."
//
// Frenzied Goblin's shape: the target is chosen as the trigger goes on the
// stack (CR 603.3d), before anyone knows whether {W} will be paid; "if you
// do" is not a reflexive trigger. The {W} is MayPay's prompt as the trigger
// resolves, and the paid branch taps the target if it is still legal
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a174adc-0536-4e12-a6fe-c44305822714",
		Name:         "Haazda Snare Squad",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Haazda Snare Squad — you may pay {W}", func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Haazda Snare Squad — pay {W} to tap target creature an opponent controls?",
						Then:     haazdaSnareSquadPaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}

// haazdaSnareSquadPaid taps the chosen creature if it is still a legal target.
var haazdaSnareSquadPaid = effects.PayUnlessBranches("pay-unless/haazda-snare-squad-tap",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.TapEachLegalTargetBody(ctx.Game, ctx.Item)
	}, nil)
