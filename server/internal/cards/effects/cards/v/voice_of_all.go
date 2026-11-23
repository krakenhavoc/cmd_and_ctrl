package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of All — Creature — Angel {2}{W}{W}, 2/2 (Dominaria Remastered,
// #5948):
//
//	"Flying
//	 As this creature enters, choose a color.
//	 This creature has protection from the chosen color."
//
// The colour is chosen as it enters (AsEnters, stored on the permanent), and
// the protection is a layer 6 grant to itself read from that choice; before
// a colour is chosen it has none, the weaker direction.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1631e9c-a993-4d44-b009-b2624d79560c",
		Name:            "Voice of All",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		AsEnters:        effects.ChooseColorAsEnters(game.ColorForProtection, "Voice of All"),
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			Label: "This creature has protection from the chosen color.",
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && source.ChosenColor != ""
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if tok := game.ProtectionFromColor(source.ChosenColor); tok != "" {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
				}
			},
		}},
	})
}
