package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Knuckleblade — Creature — Ogre Warrior {G}{U}{R}, 4/4 (Khans of
// Tarkir):
//
//	"{2}{G}: This creature gets +2/+2 until end of turn. Activate only once
//	 each turn.
//	 {2}{U}: Return this creature to its owner's hand.
//	 {R}: This creature gains haste until end of turn."
//
// Three independent rows. The pump is Basking Rootwalla's once-a-turn
// (CR 602.5b); the bounce returns the Knuckleblade itself as the ability
// resolves, doing nothing if it has already left; the haste is Roofstalker
// Wight's self grant.
//
// No simplification.
func init() {
	const pump = "{2}{G}: This creature gets +2/+2 until end of turn. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "5a4fabfe-a641-49b7-8134-9d9edeb8bf1b",
		Name:         "Savage Knuckleblade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:     pump,
				Purpose:   game.Purpose{Answers: game.AnswerPump},
				Cost:      effects.ManaCost("{2}{G}"),
				Condition: effects.OncePerTurnActivation(pump),
				Effect:    effects.ThisGetsUntilEndOfTurn(2, 2, "Savage Knuckleblade — +2/+2 until end of turn"),
			},
			{
				Label:   "{2}{U}: Return this creature to its owner's hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{2}{U}"),
				Effect:  effects.S82OldB38ReturnThisToOwnersHand,
			},
			{
				Label:   "{R}: This creature gains haste until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("haste", "Savage Knuckleblade — haste until end of turn"),
			},
		},
	})
}
