package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cant_get_counters.go — #6241: players can't get counters (CR 122.1).
// The engine half is game/cant_get_counters.go.
//
//	Static: []game.StaticAbility{PlayersCantGetCounters()}  // Solemnity: "Players can't get counters."
//	PlayerCantGetCountersWhileSourceRemains(ctx, player, label) // Suncleanser's second mode

// PlayersCantGetCounters is "Players can't get counters.": a world rule
// the permanent carries on itself (game.PlayersCantGetCounters). Every
// counter placement onto a player, of every kind, is refused while it is
// on the battlefield with its abilities.
func PlayersCantGetCounters() game.StaticAbility {
	return game.StaticAbility{
		Label:     "Players can't get counters.",
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.ExtraRestrictions |= game.PlayersCantGetCounters
		},
	}
}

// PlayerCantGetCountersWhileSourceRemains is "That player can't get
// counters for as long as this creature remains on the battlefield"
// (Suncleanser): a record that ends when the source leaves the
// battlefield (CR 611.2b), and never begins if the source already has
// (#1432).
func PlayerCantGetCountersWhileSourceRemains(ctx *Context, player uuid.UUID, label string) error {
	d, ok := DurationWhileSourceRemains(ctx, ctx.Source())
	if !ok {
		return nil
	}
	ctx.Game.PlayerCantGetCountersForDurationForEffect(ctx.Source(), player, d, label)
	return nil
}
