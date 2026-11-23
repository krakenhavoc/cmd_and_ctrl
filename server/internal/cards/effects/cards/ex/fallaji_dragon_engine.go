package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fallaji Dragon Engine — Artifact Creature — Dragon {8}, 5/5 (The
// Brothers' War, #5656): "Prototype {2}{R} — 1/3", "Flying", "{2}: This
// creature gets +1/+0 until end of turn."
//
// Prototype is the card's second face (CR 718, game/prototype.go), built
// by the deck importer from the inset; nothing here spells it out.
// Flying is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc8cca38-df2a-4cef-973c-a482505be66f",
		Name:            "Fallaji Dragon Engine",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 0, "Fallaji Dragon Engine — +1/+0"),
		}},
	})
}
