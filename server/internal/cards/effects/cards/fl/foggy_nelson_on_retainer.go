package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foggy Nelson, On Retainer — Legendary Creature — Human Advisor {2}{W},
// 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 When Foggy Nelson enters, put a +1/+1 counter on another target
//	 creature you control. It gains hexproof until end of turn."
//
// Two instructions on the same target, in order: the counter, then the
// hexproof (CR 608.2c). A target that became illegal gets neither (CR
// 608.2b). Foggy himself is never a legal pick.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Foggy Nelson — +1/+1 counter and hexproof on another creature you control",
		effects.Do(effects.Instruction(func(ctx *effects.Context) error { return effects.PlusOneCounterOnTheTarget(ctx.Game, ctx.Item) }), effects.Instruction(effects.Msc10HexproofTheLegalTarget)))
	enters.Targets = effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))
	effects.Register(effects.Spec{
		OracleID:        "734fc893-b736-448a-b851-826976031754",
		Name:            "Foggy Nelson, On Retainer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
