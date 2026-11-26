package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dictate of Karametra — Enchantment {3}{G}{G} (Journey into Nyx, S82 OLD,
// #5558):
//
//	"Flash
//	 Whenever a player taps a land for mana, that player adds one mana of any
//	 type that land produced."
//
// Heartbeat of Spring's triggered mana ability with flash.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01cb56d7-f859-46fd-b6c9-40d2c4965b69",
		Name:            "Dictate of Karametra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAPlayerTapsALandForMana(
				"Dictate of Karametra — add one mana of any type that land produced",
				effects.AddsOneManaOfAnyTypeProduced()),
		},
	})
}
