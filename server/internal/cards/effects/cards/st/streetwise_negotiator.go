package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Streetwise Negotiator — Creature — Cat Citizen {1}{G}, 0/2 (March of the
// Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target
//	 creature. If that's another creature, it gains the following ability
//	 until end of turn.)
//	 This creature assigns combat damage equal to its toughness rather than
//	 its power."
//
// Felothar's rule (Characteristic.DamageByToughnessAlways, CR 510.1c) on
// the Negotiator alone, as a layer-6 static; backup grants it as the scoped
// game.DamageByToughnessMod (Walking Bulwark's), so the recipient assigns by
// toughness without its power changing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74b631c5-9420-496e-8845-3b050d3d5340",
		Name:         "Streetwise Negotiator",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer6Ability,
			Label:     "This creature assigns combat damage equal to its toughness rather than its power.",
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.DamageByToughnessAlways = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.BackupAlso("Streetwise Negotiator", 1, "", []game.Mod{game.DamageByToughnessMod()}),
		},
	})
}
