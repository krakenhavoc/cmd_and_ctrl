package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kotis, the Fangkeeper — Legendary Creature — Zombie Warrior {1}{B}{G}{U},
// 2/1 (Tarkir: Dragonstorm, #4463):
//
//	"Indestructible
//	 Whenever Kotis deals combat damage to a player, exile the top X cards
//	 of their library, where X is the amount of damage dealt. You may cast
//	 any number of spells with mana value X or less from among them
//	 without paying their mana costs."
//
// Hazoret's Undying Fury's seam (ExileTopWithPermission.EachManaValueAtMost):
// the top X cards of the damaged player's library are exiled, with one
// free, pass-closed permission held by Kotis's controller over the cards
// that landed, capped at mana value X on each spell with no running total.
// The casts are made during the resolution (CR 608.2g), each its own
// answer until the controller declines or none is left; a card not cast,
// a land and a spell over mana value X stay in exile. An {X} in a spell's
// cost is 0 (the free cast locks it). X is the damage as the event
// recorded it, even after Kotis has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2eb2f6a2-e31a-4d3e-93ce-100394899014",
		Name:            "Kotis, the Fangkeeper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Kotis, the Fangkeeper — exile the top X cards of their library; you may cast spells with mana value X or less from among them",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ev := ctx.Trigger().Event
					limit := ev.Amount
					if limit <= 0 {
						return nil
					}
					return ExileTopWithPermission{
						From:                ev.Target,
						GrantTo:             item.Controller,
						N:                   limit,
						EachManaValueAtMost: &limit,
					}.Apply(ctx)
				}),
		},
	})
}
