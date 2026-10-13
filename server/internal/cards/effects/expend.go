package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// expend.go — "Whenever you expend N" (Bloomburrow, #3172): "You expend N
// as you spend your Nth total mana to cast spells during a turn."
//
// The count is a per-player turn tally (CastTally.ManaSpent, written as
// each spell becomes cast from its payment record, game/mutations.go),
// and the trigger fires on the cast that carried the running total across
// N: ManaSpent >= N and the total before this spell was below N
// (CastTally.Expended). So a single spell that spends the 7th through the
// 9th mana triggers expend 8 once, a later spell does not trigger it
// again, and expend 4 and expend 8 of the same turn are separate
// triggers. Mana spent on abilities does not count, and a payment the
// engine waived (a free cast on paper) counts as no mana: weaker than
// printed, never stronger.

// WheneverYouExpend is "Whenever you expend N, <effect>". The trigger
// goes on the stack above the spell whose payment crossed N, as printed
// ("as you spend", CR 603.2 puts it above the spell being cast).
func WheneverYouExpend(n int, label string, effect Effect) game.TriggeredAbility {
	return On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return g != nil && ev.Actor == source.Controller && g.CastTallyFor(ev.Actor).Expended(n)
	}, label, effect)
}
