package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tuvasa the Sunlit — Legendary Creature — Merfolk Shaman {G}{W}{U}, 1/1
// (Commander 2018, #5526):
//
//	"Tuvasa gets +1/+1 for each enchantment you control.
//	 Whenever you cast your first enchantment spell each turn, draw a
//	 card."
//
// The pump is a layer 7c modification of Tuvasa alone whose size is the
// enchantments its controller controls, counted on every recompute (CR
// 611.3a). The draw is Psemilla's trigger: the per-turn cast tally is bumped
// before the cast event, so a count of one means this is your first
// enchantment spell this turn (an enchantment creature counts; a copy and an
// opponent's spell do not).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a83d976-8087-48d7-b6f5-63efb8329e99",
		Name:         "Tuvasa the Sunlit",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			s82NeoB18SelfPT("Tuvasa gets +1/+1 for each enchantment you control.", func(g *game.Game, source *game.Card) int {
				return len(permanentsControlledByMatching(g, source.Controller, Enchantment()))
			}),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, AllOf(YouCast(Enchantment()), youCastYourFirstEnchantmentSpellEachTurn),
				"Tuvasa the Sunlit — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
