package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ring of Renewal — Artifact {5} (Fallen Empires, S82 OLD b855, #5637):
//
//	"{5}, {T}: Discard a card at random, then draw two cards."
//
// The random discard (CR 701.9b) happens first, then the draw, so a card just drawn is
// never the one discarded; an empty hand discards nothing and still draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "45964368-ea9e-415f-9faf-870c7bf550d4",
		Name:         "Ring of Renewal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}: Discard a card at random, then draw two cards.",
			Purpose: game.Purpose{Draws: 2, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
			Effect:  effects.Do(effects.DiscardCards{N: 1}, effects.DrawCards{N: 2}),
		}},
	})
}
