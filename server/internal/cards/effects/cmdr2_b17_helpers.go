package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b17_helpers.go — card-side helpers of cmdr2 batch 17
// (Squirrels, Rats and Saprolings, #5368).

// cmdr2B17SquirrelMobPT is "This creature gets +1/+1 for each other
// Squirrel on the battlefield" (Squirrel Mob).
func cmdr2B17SquirrelMobPT() game.StaticAbility {
	return game.StaticAbility{
		Label:    "Gets +1/+1 for each other Squirrel on the battlefield.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			for _, o := range g.Battlefield.Cards {
				if o.InstanceID != source.InstanceID && o.IsCreature() && o.HasSubtype(source.TextSubtype("Squirrel")) {
					n++
				}
			}
			c.Power += n
			c.Toughness += n
		},
	}
}

// cmdr2B17HasAnySubtype is the class Swarmyard Massacre names: a creature
// with any of the given types, read through its effective subtypes. The
// caller reads the type words through the spell's own text (CR 612.2).
func cmdr2B17HasAnySubtype(c game.Card, types []string) bool {
	if !c.IsCreature() {
		return false
	}
	for _, t := range types {
		if c.HasSubtype(t) {
			return true
		}
	}
	return false
}

// cmdr2B17VanishingKeyword is the Keyword of a vanishing upkeep row.
const cmdr2B17VanishingKeyword = "vanishing"

// cmdr2B17VanishingEntry is the first of vanishing's three abilities
// (CR 702.63a): "This permanent enters with N time counters on it."
func cmdr2B17VanishingEntry(n int) game.ReplacementEffect {
	r := b10EntersWithCounters(game.CounterTime, n, "Vanishing: enters with time counters on it")
	r.RowLabel = fmt.Sprintf("Vanishing %d", n)
	return r
}

// cmdr2B17VanishingTriggers are the other two (CR 702.63a): at the
// beginning of your upkeep, if this permanent has a time counter on it,
// remove a time counter from it (an intervening "if", CR 603.4, checked
// again as it resolves by removeACounterFromThis); and when the last is
// removed from it, sacrifice it. The second reads any removal, not only
// the upkeep one (an Aether Snap-style effect that takes the last
// counter off still sacrifices it).
func cmdr2B17VanishingTriggers() []game.TriggeredAbility {
	upkeep := On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
		return ByYou(ev, source, c, g) && source.Counters[game.CounterTime] > 0
	}, "Vanishing — remove a time counter", removeACounterFromThis(game.CounterTime))
	upkeep.Keyword = cmdr2B17VanishingKeyword
	last := On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return ev.Target == source.InstanceID && ev.Label == game.CounterTime &&
			ev.Amount == 0 && g.CounterCountBeforeForEffect(ev) > 0
	}, "Vanishing — sacrifice it (last time counter removed)", SacrificeThisIfStillOnBattlefield)
	return []game.TriggeredAbility{upkeep, last}
}
