package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// theros_gods.go — the Theros God clause, "As long as your devotion to
// <colour> is less than five, <this> isn't a creature." (CR 700.5,
// CR 205.1b), as one shared static.
//
// A layer-4 type-changing static on the God itself, re-read on every
// layer recompute, so the God turns on and off with the board and no
// bookkeeping. While it is off the God is an enchantment and nothing
// else: it loses the creature type and, with it, its creature types
// (notACreature).
//
// The Gods written before this file (Thassa, God of the Sea; Heliod,
// Sun-Crowned; Thassa, Deep-Dwelling) carry the same body inline; a
// new God calls this instead.
//
// Append-only.

// godUnlessDevotion is the God clause for `color` ("B", "U", …).
func godUnlessDevotion(color string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID &&
				devotionTo(g, source.Controller, color) < 5
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			notACreature(c)
		},
	}
}

// devotionToColors is CR 700.5's two-colour devotion, "your devotion to
// black and red": the number of mana symbols among the mana costs of
// permanents `controller` controls that are any of `colors`. A hybrid
// symbol of both ({B/R}) is one symbol, not one per colour, which is
// the difference from adding devotionTo(B) and devotionTo(R).
func devotionToColors(g *game.Game, controller uuid.UUID, colors ...string) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		cost, err := game.ParseCost(c.ManaCost)
		if err != nil {
			continue
		}
		for _, req := range cost.Required {
			if symbolIsAnyColor(req.Options, colors) {
				n++
			}
		}
	}
	return n
}

func symbolIsAnyColor(options, colors []string) bool {
	for _, opt := range options {
		for _, col := range colors {
			if opt == col {
				return true
			}
		}
	}
	return false
}

// godUnlessDevotionToColors is the God clause for "As long as your
// devotion to <colour> and <colour> is less than <threshold>, <this>
// isn't a creature" (Mogis, God of Slaughter: black and red, seven).
func godUnlessDevotionToColors(threshold int, colors ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer: game.Layer4Type,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID &&
				devotionToColors(g, source.Controller, colors...) < threshold
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			notACreature(c)
		},
	}
}
