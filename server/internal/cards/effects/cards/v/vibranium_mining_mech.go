package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vibranium Mining Mech — {4} Artifact — Vehicle 6/6 (Marvel Super Heroes
// Commander, #4455):
//
//	"Trample
//	 Whenever this Vehicle enters or attacks, create a tapped Vibranium
//	 token. (It's an artifact with indestructible and "{T}: Add {C}. This
//	 mana can't be spent to cast a nonartifact spell.")
//	 {2}: This Vehicle gets +1/+0 until end of turn.
//	 Crew 2"
//
// One enters-or-attacks trigger (Royal Talon Fighter Jet's
// WhenThisEntersOrAttacks) creating Dora Milaje Elite's printed Vibranium
// token tapped. The pump is Akroma's firebreathing on the Vehicle itself and
// works whether or not it is crewed. Crew is Dependable Quinjet's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1bea6d02-186e-46d8-b788-d4cc37dbb480",
		Name:            "Vibranium Mining Mech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks("Vibranium Mining Mech — create a tapped Vibranium token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.VibraniumToken()).EntersTapped(), N: 1})),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This Vehicle gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{Target: ctx.Source(), Power: 1, Label: "Vibranium Mining Mech — +1/+0"}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Vibranium Mining Mech"),
			},
		},
	})
}
