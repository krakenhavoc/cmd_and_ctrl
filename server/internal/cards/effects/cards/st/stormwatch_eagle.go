package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormwatch Eagle — Creature — Bird {3}{U}, 2/1 (Prophecy):
//
//	"Flying
//	 Sacrifice a land: Return this creature to its owner's hand."
//
// Spitting Spider's sacrifice cost with a bounce of the source. The ability
// returns the permanent it came from only while it is still that object (CR
// 400.7, CR 602.5): an Eagle that already left, or came back as a new object,
// is not bounced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "640b67d8-75a4-44d4-a2f5-7323bad4e1a8",
		Name:            "Stormwatch Eagle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.B08SacrificeALand(),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}
