package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sphinx of the Guildpact — Artifact Creature — Sphinx {7}, 5/5 (Ravnica
// Allegiance):
//
//	"Sphinx of the Guildpact is all colors.
//	 Flying
//	 Hexproof from monocolored (This creature can't be the target of monocolored
//	 spells or abilities your opponents control.)"
//
// The colors are a layer-5 static (Fallaji Wayfarer's); hexproof from
// monocolored is Niv-Mizzet, Supreme's printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b410473-1df9-4d16-9b4d-c3ca8e4635b7",
		Name:            "Sphinx of the Guildpact",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.HexproofFromMonocolored},
		Static: []game.StaticAbility{{
			Label:     "Sphinx of the Guildpact is all colors.",
			Layer:     game.Layer5Color,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Colors = []string{"W", "U", "B", "R", "G"}
			},
		}},
	})
}
