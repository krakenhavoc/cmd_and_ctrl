package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order of the Stars — Creature — Human Cleric {W}, 0/1 (Guildpact):
//
//	"Defender
//	 As this creature enters, choose a color.
//	 This creature has protection from the chosen color."
//
// The choice is made as it enters (Flickering Ward's ChooseColorAsEnters) and
// lives on the permanent; protection is a layer 6 keyword on itself from the
// chosen colour, so before a colour is chosen it has none, and a new colour is
// chosen if it re-enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "66418c65-a15f-489e-8a2d-6cee1b5d977c",
		Name:            "Order of the Stars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForProtection, "Order of the Stars"),
		Static: []game.StaticAbility{{
			Label: "This creature has protection from the chosen color.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return source != nil && target.InstanceID == source.InstanceID && source.ChosenColor != ""
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionFromColor(source.ChosenColor))
			},
		}},
	})
}
