package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artificer's Assistant — Creature — Bird {U}, 1/1 (Dominaria, #5530):
//
//	"Flying
//	 Whenever you cast a historic spell, scry 1. (Artifacts, legendaries, and Sagas are historic. To scry 1, look at the top card of your library, then you may put that card on the bottom.)"
//
// Jhoira's historic cast trigger (CR 700.6) with a scry 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6191a7eb-04b5-47e6-a31f-43f9aa91d4d7",
		Name:            "Artificer's Assistant",
		Purpose:         game.Purpose{Scry: 1},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastAHistoricSpell, "Artificer's Assistant — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
