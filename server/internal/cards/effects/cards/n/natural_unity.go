package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Unity — Conspiracy (Conspiracy: Take the Crown, S82 OLD batch 267,
// #5540):
//
//	"Hidden agenda
//	 Creatures you control with the chosen name have "At the beginning of
//	 combat on your turn, you may pay {G}. If you do, put a +1/+1 counter on
//	 this creature.""
//
// Hidden agenda (ADR 0227). A layer-6 grant of a catalog bundle (ADR 0093)
// to the creatures with the chosen name; the granted trigger is the
// creature's own, so "this creature" and "your turn" are read from it. It
// puts the counter only if the creature is still the same permanent.
//
// No simplification.
func init() {
	const grantKey = "s82-old-b267/natural-unity-counter"
	effects.Register(effects.Spec{
		OracleID:     "3deb19f6-c1a3-4bca-a6ae-df5d0b42a66e",
		Name:         "Natural Unity",
		Completeness: effects.CompletenessFull,
		HiddenAgenda: effects.HiddenAgendaOneName,
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Triggered: []game.TriggeredAbility{
				effects.AtBeginningOfYourCombat("Natural Unity — you may pay {G}; if you do, put a +1/+1 counter on this creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.MayPay{
							Chooser:  item.Controller,
							Cost:     "{G}",
							Question: "Natural Unity — pay {G} to put a +1/+1 counter on this creature?",
							Then:     naturalUnityPaid,
						}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "At the beginning of combat on your turn, you may pay {G}. If you do, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.CreaturesYouControlWithTheChosenName, grantKey)},
	})
}

var naturalUnityPaid = effects.PayUnlessBranches("pay-unless/natural-unity-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.PlusOneCounterOnThisIfStillHere(ctx)
	}, nil)
