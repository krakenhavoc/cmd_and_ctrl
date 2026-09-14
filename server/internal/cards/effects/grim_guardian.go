package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grim Guardian — Enchantment Creature — Zombie {2}{B}, 1/4 (EDHREC
// rank 3244):
//
//	"Constellation — Whenever this creature or another enchantment
//	 you control enters, each opponent loses 1 life."
//
// The enchantress deck's drain. Eidolon of Blossoms' shape — one
// ability, two conditions on one event kind: the Guardian's own
// entry (it is an enchantment and counts itself) or another
// enchantment entering under the same control
// (b30SelfOrAnotherEnchantmentYouControlEntered). Each opponent loses
// 1 as a life change, not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1f1babf-13d0-4fc4-b192-127d2d5db7f1",
		Name:         "Grim Guardian",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b30SelfOrAnotherEnchantmentYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Grim Guardian — each opponent loses 1 life (constellation)",
					func(g *game.Game, item *game.StackItem) error {
						return eachOpponentLosesLife(g, item, 1)
					})
			},
		}},
	})
}
