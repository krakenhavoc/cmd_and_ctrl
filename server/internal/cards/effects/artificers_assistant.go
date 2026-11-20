package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Artificer's Assistant — Creature — Bird {U}, 1/1 (Dominaria, #5530):
//
//	"Flying
//	 Whenever you cast a historic spell, scry 1. (Artifacts, legendaries, and Sagas are historic. To scry 1, look at the top card of your library, then you may put that card on the bottom.)"
//
// Jhoira's historic cast trigger (CR 700.6) with a scry 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6191a7eb-04b5-47e6-a31f-43f9aa91d4d7",
		Name:            "Artificer's Assistant",
		Purpose:         game.Purpose{Scry: 1},
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, youCastAHistoricSpell, "Artificer's Assistant — scry 1", Do(Scry{N: 1})),
		},
	})
}
