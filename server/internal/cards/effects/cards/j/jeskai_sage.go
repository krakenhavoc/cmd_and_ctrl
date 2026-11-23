package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jeskai Sage — Creature — Human Monk {1}{U}, 1/1 (Fate Reforged,
// #5552):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature dies, draw a card."
//
// Prowess is printed. The dies trigger is Summit Sentinel's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b79d4b17-482e-415b-a63b-1f97a55e7583",
		Name:            "Jeskai Sage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Jeskai Sage — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
