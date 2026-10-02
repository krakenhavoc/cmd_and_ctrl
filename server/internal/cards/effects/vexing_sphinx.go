package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vexing Sphinx — Creature — Sphinx, {1}{U}{U}, 4/4:
//
//	"Flying
//	 Cumulative upkeep—Discard a card. (At the beginning of your upkeep, put an age counter on this permanent, then sacrifice it unless you pay its upkeep cost for each age counter on it.)
//	 When this creature dies, draw a card for each age counter on it."
//
// Cumulative upkeep with a discard (CR 702.24a, ADR 0108 §5): with N age
// counters the controller discards N cards, chosen and paid together, or
// sacrifices the Sphinx. The dies trigger counts the age counters it had as
// it last existed on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c2affd47-2404-44b0-a571-e64e948130eb",
		Name:            "Vexing Sphinx",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeepPaying("Vexing Sphinx — cumulative upkeep: discard a card", DiscardPayment(1), "", ""),
			WhenThisDies("Vexing Sphinx — draw a card for each age counter on it", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, _ := ctx.TriggeringPermanent()
				return DrawCards{Player: item.Controller, N: info.Counters[game.CounterAge]}.Apply(ctx)
			}),
		},
	})
}
