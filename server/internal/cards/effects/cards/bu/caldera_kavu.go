package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caldera Kavu — Creature — Kavu {2}{R}, 2/2 (Planeshift, S82 OLD #5615):
//
//	"{1}{B}: This creature gets +1/+1 until end of turn.
//	 {G}: This creature becomes the color of your choice until end of turn."
//
// The pump is the shared ThisGetsUntilEndOfTurn body. The colour is
// Spiritmonger's row: asked as the ability resolves, then set until end of
// turn so the Kavu is that colour and no other (CR 105.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7cdcfbdf-5ec7-451b-8c96-f19d75a4d76c",
		Name:         "Caldera Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
				Cost:    effects.ManaCost("{1}{B}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Caldera Kavu — +1/+1 until end of turn"),
			},
			{
				Label:   "{G}: This creature becomes the color of your choice until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerProtect},
				Cost:    effects.ManaCost("{G}"),
				Effect:  effects.OldB638BecomesChosenColorUntilEOT("Caldera Kavu", calderaKavuColorChosen),
			},
		},
	})
}

// calderaKavuColorChosen makes the Kavu the chosen colour until end of turn,
// if it is still the permanent that asked. An on-disk key.
var calderaKavuColorChosen = effects.OldB638ColorChosenTail("caldera-kavu/color-chosen", "Caldera Kavu")
