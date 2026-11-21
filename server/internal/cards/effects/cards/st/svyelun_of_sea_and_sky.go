package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Svyelun of Sea and Sky — Legendary Creature — Merfolk God {1}{U}{U}, 3/4:
//
//	"Svyelun has indestructible as long as you control at least two
//	 other Merfolk.
//	 Whenever Svyelun attacks, draw a card.
//	 Other Merfolk you control have ward {1}."
//
// The indestructible clause is a live layer 6 grant (CR 611.3a) that
// counts the other Merfolk on every recompute. Ward is a triggered
// ability (CR 702.21a), so the grant is WardGranted over the other
// Merfolk you control, as Long River Lurker does for Frogs.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "872bebff-720d-459b-8f86-fb9129a8d1fb",
		Name:         "Svyelun of Sea and Sky",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.Cmdr2B05SelfKeywordWhileOtherMerfolk(2, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Svyelun of Sea and Sky — draw a card", effects.Do(effects.DrawCards{N: 1})),
			effects.WardGranted(effects.WardMana("{1}"), "Svyelun of Sea and Sky — other Merfolk you control have ward {1}",
				effects.Cmdr2B05MerfolkFilter(true, true).Matches),
		},
	})
}
