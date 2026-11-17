package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dawnglade Regent — Creature — Elk {5}{G}{G}, 8/8 (Commander Legends,
// #5507):
//
//	"When this creature enters, you become the monarch.
//	 As long as you're the monarch, permanents you control have hexproof."
//
// The enters trigger is every monarch card's. The hexproof is a layer-6
// grant over every permanent its controller controls, the Regent included
// (the ruling), read through YoureTheMonarch on each recompute, which
// EventMonarchChanged invalidates: the moment the crown moves to another
// player, your permanents can be targeted again.
//
// No simplification.
func init() {
	grant := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && YoureTheMonarch(g, source.Controller)
	}, "hexproof")
	grant.Label = "As long as you're the monarch, permanents you control have hexproof."
	Register(Spec{
		OracleID:     "bb1c6266-5e25-4a5b-b15a-faeda3ce04dc",
		Name:         "Dawnglade Regent",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Dawnglade Regent"),
		},
		Static: []game.StaticAbility{grant},
	})
}
