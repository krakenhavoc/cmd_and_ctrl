package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Abomination of Llanowar — Legendary Creature — Elf Horror {1}{B}{G}, */*
// (Kaldheim Commander, #5506):
//
//	"Vigilance; menace
//	 Abomination of Llanowar's power and toughness are each equal to the
//	 number of Elves you control plus the number of Elf cards in your
//	 graveyard."
//
// A characteristic-defining ability (layer 7a), read live: the Elves you
// control (itself counted, through the layered view) plus the Elf cards in
// your graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5146ba3f-2c5f-4a86-95d7-a34ce1e842b0",
		Name:            "Abomination of Llanowar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "menace"},
		Static: []game.StaticAbility{{
			Label:     "Abomination of Llanowar's power and toughness are each equal to the number of Elves you control plus the number of Elf cards in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.S82OldB074ElvesYouControl(g, source.Controller, source) + effects.S82OldB074ElfCardsInGraveyard(g, source.Controller, source)
				c.Power, c.Toughness = n, n
			},
		}},
	})
}
