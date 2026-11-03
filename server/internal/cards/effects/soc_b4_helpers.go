package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// soc_b4_helpers.go — shared bodies for the fourth Secrets of
// Strixhaven Commander batch (#4834).

// socB4ThisBecomesPrepared is "this creature becomes prepared" as a
// trigger's effect (CR 722.3a): the permanent the ability came from,
// which BecomePrepared leaves alone when it has left the battlefield
// or is already prepared.
func socB4ThisBecomesPrepared(g *game.Game, item *game.StackItem) error {
	return BecomePrepared{Target: item.SourceCardID}.Apply(NewContext(g, item))
}

// socB4TriggeringSpellManaValue is "that spell's mana value" for a cast
// or copy trigger: the spell on the stack with its announced X, or as it
// last existed there when it left before the trigger resolved (CR
// 608.2h). A spell found in neither falls back to the card wherever it
// is now (triggeringSpellManaValue).
func socB4TriggeringSpellManaValue(g *game.Game, item *game.StackItem) int {
	if item.Trigger == nil {
		return 0
	}
	if mv, ok := g.LastKnownSpellManaValueForEffect(item.Trigger.Event.CardID); ok {
		return mv
	}
	return triggeringSpellManaValue(g, item)
}
