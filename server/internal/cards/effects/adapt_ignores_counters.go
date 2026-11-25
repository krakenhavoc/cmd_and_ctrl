package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// adapt_ignores_counters.go — "The next time target creature adapts this
// turn, it adapts as though it had no +1/+1 counters on it" (Biomancer's
// Familiar, #6239). The engine half is game/adapt_ignores_counters.go;
// every adapt goes through AdaptN, which spends the record.

// NextAdaptIgnoresCounters is the clause on the item's legal targets:
// each gets a one-shot record, pinned to the object until cleanup.
//
//	Targets: TargetCreature("target creature"),
//	Effect:  NextAdaptIgnoresCounters("Biomancer's Familiar"),
func NextAdaptIgnoresCounters(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			g.AdaptIgnoresCountersThisTurnForEffect(item.SourceCardID, t.ID,
				name+" — the next time it adapts this turn, it adapts as though it had no +1/+1 counters on it")
		}
		return nil
	}
}
