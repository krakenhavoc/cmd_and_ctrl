package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire-Belly Changeling — Creature — Shapeshifter {1}{R}, 1/1 (Lorwyn, S82
// OLD b548, #5593):
//
//	"Changeling (This card is every creature type.)
//	 {R}: This creature gets +1/+0 until end of turn. Activate no more than
//	 twice each turn."
//
// Roterothopter's limit: activations are counted as they are announced
// (AtMostNActivationsPerTurn), so a third activation is refused even while
// the first two are still on the stack.
//
// No simplification.
func init() {
	const label = "{R}: This creature gets +1/+0 until end of turn. Activate no more than twice each turn."
	effects.Register(effects.Spec{
		OracleID:        "65c1cac7-87d1-4687-86ba-f7c63c4f8549",
		Name:            "Fire-Belly Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1}},
			Cost:      effects.ManaCost("{R}"),
			Condition: effects.AtMostNActivationsPerTurn(label, 2),
			Effect:    effects.ThisGetsUntilEndOfTurn(1, 0, "Fire-Belly Changeling — +1/+0 until end of turn"),
		}},
	})
}
