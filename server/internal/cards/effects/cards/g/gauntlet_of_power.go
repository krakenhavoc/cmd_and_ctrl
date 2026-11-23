package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gauntlet of Power — Artifact {5} (Dominaria Remastered):
//
//	"As this artifact enters, choose a color.
//	 Creatures of the chosen color get +1/+1.
//	 Whenever a basic land is tapped for mana of the chosen color, its
//	 controller adds an additional one mana of that color."
//
// Caged Sun's two halves, widened. The anthem reaches every creature of
// the chosen colour, whoever controls it (gauntletOfPowerAnthem). The
// mana half is a CR 605.1b triggered mana ability like Mana Flare's: it
// fires for any player's BASIC land (read off the supertype as it was
// when tapped) that produced the chosen colour, and the extra mana goes
// to the player whose pool the land's mana landed in, "its controller".
// A land that produced the colour twice still gets one extra. Before a
// colour is chosen both halves do nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1182591d-8660-4a2e-8d77-2fc544b02b66",
		Name:         "Gauntlet of Power",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Gauntlet of Power"),
		Static:       []game.StaticAbility{gauntletOfPowerAnthem()},
		ManaTriggers: []game.ManaTrigger{{
			Label: "Gauntlet of Power — add an additional one mana of the chosen color",
			AppliesTo: func(prod game.ManaProduced, source *game.Card, g *game.Game) bool {
				if !effects.IsBasicLand(prod.Source) {
					return false
				}
				chosen := g.ChosenColorOf(source.InstanceID)
				if chosen == "" {
					return false
				}
				for _, c := range prod.Colors {
					if c == chosen {
						return true
					}
				}
				return false
			},
			Produced: effects.AddsOneManaOfTheChosenColor(),
		}},
	})
}

// gauntletOfPowerAnthem is "Creatures of the chosen color get +1/+1":
// layer 7c over every creature, not only the controller's.
func gauntletOfPowerAnthem() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Creatures of the chosen color get +1/+1.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return source.ChosenColor != "" && target.IsCreature() && target.HasColor(source.ChosenColor)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Power++
			c.Toughness++
		},
	}
}
