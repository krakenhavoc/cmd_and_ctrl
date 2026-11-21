package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stillmoon Cavalier — Creature — Zombie Knight {1}{W/B}{W/B}, 2/1 (Eventide, #5590):
//
//	"Protection from white and from black
//	 {W/B}: This creature gains flying until end of turn.
//	 {W/B}: This creature gains first strike until end of turn.
//	 {W/B}{W/B}: This creature gets +1/+0 until end of turn."
//
// Animar's two protections and Order of Leitbur's rows, each paid with
// white or black mana.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c430b5a-83b6-4f4f-b1d3-5737e1f61947",
		Name:            "Stillmoon Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white", "protection from black"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W/B}: This creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    effects.ManaCost("{W/B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Stillmoon Cavalier — flying until end of turn"),
			},
			{
				Label:   "{W/B}: This creature gains first strike until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{W/B}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("first strike", "Stillmoon Cavalier — first strike until end of turn"),
			},
			{
				Label:   "{W/B}{W/B}: This creature gets +1/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{W/B}{W/B}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Stillmoon Cavalier — +1/+0 until end of turn"),
			},
		},
	})
}
