package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frenzied Arynx — Creature — Cat Beast {2}{R}{G}, 3/3:
//
//	"Riot (This creature enters with your choice of a +1/+1 counter or
//	 haste.)
//	 Trample
//	 {4}{R}{G}: This creature gets +3/+0 until end of turn."
//
// Riot is the engine's keyword (ADR 0109 §10); the pump is the shared
// "this creature gets +N/+M until end of turn" body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "772873f4-74b9-44da-bd69-e4c8de797a26",
		Name:            "Frenzied Arynx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRiot, "trample"},
		Activated: []ActivatedAbility{{
			Label:   "{4}{R}{G}: This creature gets +3/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{4}{R}{G}"),
			Effect:  thisGetsUntilEndOfTurn(3, 0, "Frenzied Arynx — +3/+0"),
		}},
	})
}
