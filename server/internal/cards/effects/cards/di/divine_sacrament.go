package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Divine Sacrament — Enchantment {1}{W}{W} (Dominaria Remastered, #5948):
//
//	"White creatures get +1/+1.
//	 Threshold — White creatures get an additional +1/+1 as long as there
//	 are seven or more cards in your graveyard."
//
// Two layer 7c statics on every white creature, whoever controls it. The
// colour is read through the layers (a creature that is white because of an
// effect counts), and the threshold one is read live against the
// enchantment's controller's graveyard.
//
// No simplification.
func init() {
	isWhite := func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.HasColor(source.TextColor("W"))
	}
	effects.Register(effects.Spec{
		OracleID:     "92bd7fe2-bb70-40fb-984b-26e461347c56",
		Name:         "Divine Sacrament",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "White creatures get +1/+1.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: isWhite,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
			{
				Label:    "Threshold — White creatures get an additional +1/+1 while there are seven or more cards in your graveyard.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return isWhite(target, g, source) && effects.FraThresholdMet(g, source.Controller)
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power++
					c.Toughness++
				},
			},
		},
	})
}
