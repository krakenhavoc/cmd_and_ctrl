package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suspicious Stowaway // Seafaring Werewolf — {1}{U} Creature — Human
// Rogue Werewolf 1/1 // Creature — Werewolf 2/1 (#2586, ADR 0132):
//
//	Front: "This creature can't be blocked.
//	        Whenever this creature deals combat damage to a player, draw a
//	        card, then discard a card.
//	        Daybound"
//	Back:  "This creature can't be blocked.
//	        Whenever this creature deals combat damage to a player, draw a
//	        card.
//	        Nightbound"
//
// Silent Hallcreeper's evasion and the usual combat-damage trigger; the
// front's loot is Looter il-Kor's draw-then-discard.
//
// No simplification.
func init() {
	const oracle = "da412474-1da8-409b-bd7b-87cbfc239fb0"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Suspicious Stowaway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Suspicious Stowaway — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return effects.B16DrawThenDiscard(g, item, 1) }),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Seafaring Werewolf",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Seafaring Werewolf — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
