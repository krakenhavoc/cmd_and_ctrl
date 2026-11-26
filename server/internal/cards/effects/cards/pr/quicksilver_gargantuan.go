package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quicksilver Gargantuan — Creature — Shapeshifter {5}{U}{U}, 7/7 (Scars of
// Mirrodin, #5577):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it's 7/7."
//
// Clone's replacement (EntersAsCopyOf, CR 614.1c, CR 707.2) with an except
// clause that sets the copiable power and toughness to 7/7 (CR 707.9b). Not
// targeting. Declined, or with nothing to copy, it enters as itself, a 7/7.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c30ffe3f-ec1b-4af5-b817-7ef9a1c5d2f3",
		Name:         "Quicksilver Gargantuan",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf(
				"Quicksilver Gargantuan",
				effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.SetPowerToughness(7, 7)
				},
			),
		},
	})
}
