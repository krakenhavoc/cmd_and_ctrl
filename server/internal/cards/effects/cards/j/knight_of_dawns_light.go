package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Knight of Dawn's Light — Creature — Human Knight {1}{W}, 2/2 (Dominaria
// United):
//
//	"First strike
//	 If you would gain life, you gain that much life plus 1 instead.
//	 {1}{W}: This creature gets +1/+1 until end of turn."
//
// Angel of Vitality's life-gain replacement (CR 614.1a), which reaches every
// gain of yours, plus a firebreathing-style pump on the card itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "801e55c7-f8a1-4a90-81cb-3ad7022290ba",
		Name:            "Knight of Dawn's Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Replacements: []game.ReplacementEffect{
			effects.LifeGainBecomes{
				Amount: func(n int) int { return n + 1 },
				Scope:  effects.LifeGainByController,
				Label:  "Knight of Dawn's Light: gain that much life plus 1",
			}.Build(),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{W}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Knight of Dawn's Light — +1/+1 until end of turn"),
		}},
	})
}
