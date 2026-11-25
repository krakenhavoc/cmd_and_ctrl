package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Malakir Familiar — Creature — Bat {2}{B}, 2/1 (Jumpstart, #5510):
//
//	"Flying, deathtouch
//	 Whenever you gain life, this creature gets +1/+1 until end of turn."
//
// Rodolf Duskbringer's life-gain trigger with a pump: one trigger per life
// gain event, however much life it was.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12eeabe1-a02c-42a3-aa87-6774c4de7093",
		Name:            "Malakir Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Malakir Familiar — it gets +1/+1 until end of turn",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Malakir Familiar — +1/+1")),
		},
	})
}
