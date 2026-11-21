package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sparktongue Dragon — Creature — Dragon {3}{R}{R}, 3/3 (Foundations
// Jumpstart):
//
//	"Flying
//	 When this creature enters, you may pay {2}{R}. When you do, it deals 3
//	 damage to any target."
//
// Boilerbilges Ripper's reflexive shape with a mana payment: the offer is
// MayPay, and only a paid cost creates the reflexive trigger (CR 603.12),
// whose target is chosen then. The damage comes from the Dragon (last-known
// information if it has left, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1817acd5-5ed0-47b8-8e56-4715a488867e",
		Name:            "Sparktongue Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(sparktongueDragonLabel, sparktongueDragonOffer),
		},
	})
}

const sparktongueDragonLabel = "Sparktongue Dragon — you may pay {2}{R}; when you do, 3 damage to any target"

func sparktongueDragonOffer(g *game.Game, item *game.StackItem) error {
	return effects.MayPay{
		Chooser:  item.Controller,
		Cost:     "{2}{R}",
		Question: sparktongueDragonLabel + "?",
		Then:     sparktongueDragonPaid,
	}.Apply(effects.NewContext(g, item))
}

var sparktongueDragonPaid = effects.PayUnlessBranches("pay-unless/sparktongue-dragon-damage",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label:  "Sparktongue Dragon — 3 damage to any target",
			Body:   sparktongueDragonDamageBody,
			Params: game.EffectParams{Amount: 3},
		}.Apply(ctx)
	}, nil)

var sparktongueDragonDamageBody = game.ReflexiveBody("sparktongue-dragon/damage", effects.BreechesBlastEffect, effects.ConstTargets(effects.TargetAny))
