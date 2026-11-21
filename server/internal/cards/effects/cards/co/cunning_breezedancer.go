package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cunning Breezedancer — Creature — Dragon {4}{W}{U}, 4/4 (Dragons of Tarkir,
// #5551):
//
//	"Flying
//	 Whenever you cast a noncreature spell, this creature gets +2/+2 until end of turn."
//
// Frenzied Devils' cast trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ba997e44-fc80-4463-ad3f-6497be26df0f",
		Name:            "Cunning Breezedancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(), "Cunning Breezedancer — gets +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Cunning Breezedancer — +2/+2")),
		},
	})
}
