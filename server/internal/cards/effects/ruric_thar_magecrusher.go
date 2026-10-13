package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ruric Thar, Magecrusher — Legendary Creature — Ogre Warrior
// {5}{G}{G}, 7/7:
//
//	"This spell can't be countered.
//	 Reach, vigilance, trample
//	 Ruric Thar has hexproof as long as they haven't dealt combat
//	 damage yet."
//
// The conditional hexproof is a layer-6 self-grant behind the
// HasntDealtCombatDamage gate (#3172, ADR 0071's amendment of
// 2026-10-13). The gate reads Card.DealtCombatDamage, which the engine
// stamps when combat damage this permanent was the source of actually
// lands (emitDealDamageLocked — damage fully prevented is not dealt) and
// which schedules the layer pass, so the hexproof is gone from the next
// priority on. It stays gone until Ruric Thar leaves the battlefield and
// returns as a new object (CR 400.7).
func init() {
	Register(Spec{
		OracleID:        "453b2b3c-7df5-4515-9565-82f603cf451b",
		Name:            "Ruric Thar, Magecrusher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance", "trample"},
		CantBeCountered: true,
		Static: []game.StaticAbility{{
			Label:      "Has hexproof until it deals combat damage.",
			Layer:      game.Layer6Ability,
			ActiveWhen: game.HasntDealtCombatDamage(),
			AppliesTo:  selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				appendKeywordsTo(c, []string{"hexproof"})
			},
		}},
	})
}
