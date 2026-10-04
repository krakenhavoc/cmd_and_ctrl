package effects

import (
	"slices"

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

// devotionToColors is CR 700.5's devotion to a COMBINATION of colours:
// the mana symbols among the mana costs of permanents `controller`
// controls that are any of `colors`. A symbol counts once however many
// of the colours it is, so {R/G} adds one to a devotion to red and
// green, not two (the Xenagos, God of Revels ruling). An unparseable
// or empty mana cost contributes nothing.
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
			if slices.ContainsFunc(req.Options, func(opt string) bool { return slices.Contains(colors, opt) }) {
				n++
			}
		}
	}
	return n
}

// godUnlessDevotionTo is the God clause with any threshold and any
// colours: "As long as your devotion to red and green is less than
// seven, Xenagos isn't a creature" is godUnlessDevotionTo(7, "R", "G").
func godUnlessDevotionTo(threshold int, colors ...string) game.StaticAbility {
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
