package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galecaster Colossus — Creature — Giant Wizard {5}{U}{U}, 5/6 (Commander
// 2017, #5533):
//
//	"Tap an untapped Wizard you control: Return target nonland permanent
//	 you don't control to its owner's hand."
//
// The cost is Zombie Trailblazer's TapAnUntapped: not the {T} symbol
// (CR 302.6), so a Wizard that arrived this turn may pay it, and the
// Colossus, itself a Wizard, may pay with itself. The bounce is Temporal
// Adept's over Gust of Wind's "nonland permanent you don't control"; a
// target gone or illegal by resolution leaves nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8379d8f-df17-4473-938d-7ac198d17980",
		Name:         "Galecaster Colossus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap an untapped Wizard you control: Return target nonland permanent you don't control to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    TapAnUntapped("an untapped Wizard you control", HasSubtype("Wizard")),
			Targets: TargetPermanent("target nonland permanent you don't control", And(Nonland(), OpponentControls())),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
