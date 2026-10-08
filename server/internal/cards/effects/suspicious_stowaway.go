package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        oracle,
		Name:            "Suspicious Stowaway",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"daybound"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Suspicious Stowaway — draw a card, then discard a card",
				func(g *game.Game, item *game.StackItem) error { return b16DrawThenDiscard(g, item, 1, 1) }),
		},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Seafaring Werewolf",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"nightbound"},
		Static:          []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Seafaring Werewolf — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
