package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Exquisite Blood — Enchantment {4}{B} (EDHREC rank 516):
//
//	"Whenever an opponent loses life, you gain that much life."
//
// Sanguine Bond's other half. "Loses life" is two event kinds in
// this engine — see b04OpponentLostLife — because a Bolt or a combat
// hit writes the life total directly and emits only EventDealDamage,
// while a drain emits EventChangeLife with a negative delta. Both
// count; nothing emits both for one loss, so nothing is doubled. The
// amount is captured by value in Build.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8f933fae-6c0c-42d7-a817-14760d8285cd",
		Name:         "Exquisite Blood",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventChangeLife, game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b04OpponentLostLife(ev, source.Controller, g)
				return ok
			},
			Key: "Exquisite Blood — you gain that much life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := b04OpponentLostLife(ev, source.Controller, g)
				item := game.NewTriggeredItem(source, "Exquisite Blood — you gain that much life", nil)
				item.Params.Amount = amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GainLife{Player: item.Controller, Amount: item.Params.Amount}.Apply(NewContext(g, item))
			},
		}},
	})
}
