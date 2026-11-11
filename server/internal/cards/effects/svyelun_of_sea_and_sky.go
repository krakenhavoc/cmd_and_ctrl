package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "872bebff-720d-459b-8f86-fb9129a8d1fb",
		Name:         "Svyelun of Sea and Sky",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			cmdr2B05SelfKeywordWhileOtherMerfolk(2, "indestructible"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Svyelun of Sea and Sky — draw a card", Do(DrawCards{N: 1})),
			WardGranted(WardMana("{1}"), "Svyelun of Sea and Sky — other Merfolk you control have ward {1}",
				cmdr2B05MerfolkFilter(true, true).Matches),
		},
	})
}
