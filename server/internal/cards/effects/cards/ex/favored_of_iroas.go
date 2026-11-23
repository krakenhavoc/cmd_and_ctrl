package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favored of Iroas — Creature — Human Soldier {2}{W}, 2/2:
//
//	"Constellation — Whenever an enchantment you control enters, this
//	 creature gains double strike until end of turn."
//
// Constellation as in Archon of Sun's Grace, one trigger per enchantment; the
// keyword is granted to the source through thisCreatureUntilEOT, which does
// nothing if it has left the battlefield.
//
// No simplification.
func init() {
	const label = "Favored of Iroas — gains double strike until end of turn (constellation)"
	effects.Register(effects.Spec{
		OracleID:     "5df784c0-3adc-4613-ba30-905ec8cf66f9",
		Name:         "Favored of Iroas",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && c.IsEnchantment()
				}, label, effects.ThisCreatureUntilEOT(label, 0, 0, "double strike")),
				game.Purpose{Pump: &game.Pump{Keywords: []string{"double strike"}}}),
		},
	})
}
