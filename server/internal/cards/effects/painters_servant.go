package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Painter's Servant — Artifact Creature — Scarecrow {2}, 1/3:
//
//	"As this creature enters, choose a color.
//	 All cards that aren't on the battlefield, spells, and permanents are
//	 the chosen color in addition to their other colors."
//
// The colour is stored as the permanent enters (Card.ChosenColor). The second
// line is Mycosynth Lattice's two statics with a layer 5 add in place of the
// set: one over every permanent, one over every spell and every card in a
// hand, library, graveyard, exile or the command zone (zone_layers.go, #3862).
// Until a colour is chosen the statics add nothing, the weaker reading.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "baf93873-35e6-4bf2-bdcc-78a5206422fb",
		Name:         "Painter's Servant",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForBenefit, "Painter's Servant"),
		Static: []game.StaticAbility{
			{
				Layer: game.Layer5Color,
				Label: "All permanents are the chosen color in addition to their other colors.",
				AppliesTo: func(_ *game.Card, _ *game.Game, source *game.Card) bool {
					return source.ChosenColor != ""
				},
				Apply: painterAddsChosenColor,
			},
			{
				Layer:        game.Layer5Color,
				Label:        "All cards that aren't on the battlefield and spells are the chosen color in addition to their other colors.",
				AffectsZones: OffTheBattlefield(),
				AppliesTo: func(_ *game.Card, _ *game.Game, source *game.Card) bool {
					return source.ChosenColor != ""
				},
				Apply: painterAddsChosenColor,
			},
		},
	})
}

// painterAddsChosenColor adds the source's chosen colour to the object's
// colours, once.
func painterAddsChosenColor(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
	for _, col := range c.Colors {
		if col == source.ChosenColor {
			return
		}
	}
	c.Colors = append(append([]string(nil), c.Colors...), source.ChosenColor)
}
