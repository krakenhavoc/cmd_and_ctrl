package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weatherlight Compleated — Legendary Artifact — Vehicle {2}, 5/5 (Dominaria
// United, #5487):
//
//	"Flying
//	 As long as Weatherlight Compleated has four or more phyresis counters on
//	 it, it's a Phyrexian creature in addition to its other types.
//	 Whenever a creature you control dies, put a phyresis counter on
//	 Weatherlight Compleated. Then draw a card if it has seven or more
//	 phyresis counters on it. If it doesn't, scry 1."
//
// It has no crew cost: it becomes a creature only through its own counters
// (a layer 4 type change gated on the counter count, CR 613.1d), and keeps
// its printed 5/5. The trigger also sees Weatherlight Compleated itself die
// (CR 603.10a); it then cannot take the counter, and the "then" reads the
// counters it last had (CR 608.2h), so a death with seven counters
// still draws and one with fewer scries.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "060a7d77-61a8-4964-8fb6-35bd9a999517",
		Name:            "Weatherlight Compleated",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1, Draws: 1},
		Static: []game.StaticAbility{{
			Label: "As long as Weatherlight Compleated has four or more phyresis counters on it, it's a Phyrexian creature in addition to its other types.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.Counters[weatherlightCounter] >= 4
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					c.Types = append(c.Types, "Creature")
				}
				if !slices.Contains(c.Subtypes, "Phyrexian") {
					c.SetSubtypes(append(slices.Clone(c.Subtypes), "Phyrexian"))
				}
			},
		}},
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies(
				"Weatherlight Compleated — put a phyresis counter on it, then draw a card if it has seven or more, otherwise scry 1",
				weatherlightCompleatedDies),
		},
	})
}

const weatherlightCounter = "phyresis"

func weatherlightCompleatedDies(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	src := ctx.Source()
	if onBattlefield(g, src) {
		if err := (AddCounter{Target: src, Kind: weatherlightCounter, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if weatherlightCounters(ctx, src) >= 7 {
		return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
	}
	return Scry{Player: item.Controller, N: 1}.Apply(ctx)
}

// weatherlightCounters is the phyresis counters on the ship now, or, once it
// has left the battlefield, as it last existed: the dying creature the trigger
// names when that is the ship itself (CR 603.10a), else the last-known record.
func weatherlightCounters(ctx *Context, id uuid.UUID) int {
	g := ctx.Game
	if onBattlefield(g, id) {
		if c, ok := g.LookupCardForEffect(id); ok {
			return c.Counters[weatherlightCounter]
		}
	}
	if left, ok := ctx.TriggeringPermanent(); ok && ctx.Item.Trigger != nil && ctx.Item.Trigger.Event.CardID == id {
		return left.Counters[weatherlightCounter]
	}
	return g.LastKnownCountersForEffect(id)[weatherlightCounter]
}
