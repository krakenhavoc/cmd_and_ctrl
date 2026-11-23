package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Truefire Paladin — Creature — Human Knight {R}{W}, 2/2 (Gatecrash, #5565):
//
//	"Vigilance
//	 {R}{W}: This creature gets +2/+0 until end of turn.
//	 {R}{W}: This creature gains first strike until end of turn."
//
// Vigilance is the engine keyword; the two abilities are Freejam
// Regent's pump and Burning Shield Askari's grant, each pinned to this
// creature until cleanup.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a8a239c-7cc7-48d5-b34b-60b0da4e4b20",
		Name:            "Truefire Paladin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}{W}: This creature gets +2/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{R}{W}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Truefire Paladin — +2/+0 until end of turn"),
			},
			{
				Label:   "{R}{W}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"first strike"}}},
				Cost:    effects.ManaCost("{R}{W}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Truefire Paladin — first strike until end of turn"),
			},
		},
	})
}
