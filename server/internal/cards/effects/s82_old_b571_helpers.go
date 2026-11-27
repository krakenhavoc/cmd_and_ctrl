package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b571_helpers.go — shared shapes of the Guildpact creatures of S82
// OLD batch 571 (#5599).

// OldB571BloodthirstX is "Bloodthirst X" (CR 702.54b): the permanent enters
// with X +1/+1 counters on it, where X is the total damage its controller's
// opponents have been dealt this turn, from any source, combat or not. Damage
// dealt, not life lost: a drain does not count and a prevented point was never
// dealt. The total is read as the permanent enters, against the controller it
// enters under. With none dealt no counter is placed.
func OldB571BloodthirstX(name string) game.ReplacementEffect {
	return game.ReplacementEffect{
		Label:     name + " — bloodthirst X",
		RowLabel:  "Bloodthirst X: enters with X +1/+1 counters on it, where X is the total damage your opponents have been dealt this turn.",
		EntryOnly: true,
		Watches:   []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				GapB82DamageDealtToOpponents(g, src.Controller) > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, GapB82DamageDealtToOpponents(g, src.Controller))
			return nil
		},
	}
}

// OldB571EnchantedSelf is "as long as it's enchanted" for the creature that
// carries it: the target is the source itself and has an Aura attached.
func OldB571EnchantedSelf(target *game.Card, g *game.Game, source *game.Card) bool {
	return source != nil && target.InstanceID == source.InstanceID && IsEnchantedByAura(g, target.InstanceID)
}
