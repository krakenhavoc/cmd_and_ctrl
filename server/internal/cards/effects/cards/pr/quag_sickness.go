package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Quag Sickness — Enchantment — Aura {2}{B} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1 for each Swamp you control."
//
// Granite Grip's per-land count, as a shrink: the Swamps are counted on
// every layer pass, and they are the ones the Aura's controller controls,
// not the enchanted creature's controller (CR 109.5). A creature it takes
// to 0 toughness dies to the state-based check (CR 704.5f). "Swamp" is a
// word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c7f2c629-0d3a-4556-951e-6bc5d74fd7bc",
		Name:         "Quag Sickness",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(-1, -1, m11SwampsTheAurasControllerControls),
		},
	})
}

// m11SwampsTheAurasControllerControls is Quag Sickness's "each Swamp you
// control": the Aura's controller's Swamps, read through the card's text.
func m11SwampsTheAurasControllerControls(g *game.Game, source *game.Card) int {
	swamp := source.TextSubtype("Swamp")
	return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtype(swamp) })
}
