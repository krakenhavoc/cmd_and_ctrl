package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Palladia-Mors, the Ruiner — Legendary Creature — Elder Dragon {3}{R}{G}{W}, 6/6
// (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"Flying, vigilance, trample
//	 Palladia-Mors has hexproof as long as it hasn't dealt damage yet."
//
// Karakyk Guardian's gate: the hexproof exists until Card.DealtDamage is
// stamped, when damage this permanent was the source of actually lands, and
// stays gone until it leaves the battlefield and returns as a new object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c2b10ba-1930-4da8-9570-c89b2e239e3e",
		Name:            "Palladia-Mors, the Ruiner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "trample"},
		Static: []game.StaticAbility{{
			Label:      "Has hexproof as long as it hasn't dealt damage yet.",
			Layer:      game.Layer6Ability,
			ActiveWhen: game.HasntDealtDamage(),
			AppliesTo:  effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"hexproof"})
			},
		}},
	})
}
