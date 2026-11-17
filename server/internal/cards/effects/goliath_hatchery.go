package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Goliath Hatchery — Enchantment {4}{G}{G} (Phyrexia: All Will Be One,
// #5479):
//
//	"When this enchantment enters, create two 3/3 green Phyrexian Beast
//	 creature tokens with toxic 1. (Players dealt combat damage by them also
//	 get a poison counter.)
//	 Corrupted — At the beginning of your upkeep, if an opponent has three
//	 or more poison counters, choose a creature you control, then draw
//	 cards equal to its total toxic value."
//
// The enter trigger makes both Beasts in one creation. Corrupted is an
// intervening "if" (CR 603.4): checked as the upkeep begins and again on
// resolution (seedcoreCorrupted, "an opponent has three or more poison
// counters"). The choice is made on resolution, from the creatures you
// control then, and the draw reads the chosen creature's total toxic value
// (CR 702.164b: every toxic ability it has, summed) as it is chosen; a
// creature with no toxic, or no creature to choose, draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c4d65bcd-f6c5-40c3-825d-06b5183e7dbe",
		Name:         "Goliath Hatchery",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goliath Hatchery — create two 3/3 green Phyrexian Beast tokens with toxic 1",
				func(g *game.Game, item *game.StackItem) error { return s82OneB20CreateBeasts(g, item, 2) }),
			InterveningIf(CondBoard{
				Name: "an opponent has three or more poison counters",
				Test: func(ctx *Context) bool { return seedcoreCorrupted(ctx.Game, ctx.Controller(), uuid.Nil) },
			}, AtYourUpkeep("Goliath Hatchery — corrupted: choose a creature you control, then draw cards equal to its total toxic value",
				goliathHatcheryChoose)),
		},
	})
}

func goliathHatcheryChoose(g *game.Game, item *game.StackItem) error {
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx := NewContext(g, item)
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    item.Controller,
		FromPlayer: item.Controller,
		Source:     ctx.Source(),
		Question:   "Goliath Hatchery — choose a creature you control; draw cards equal to its total toxic value",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    goliathHatcheryDraw,
	})
	return nil
}

// goliathHatcheryDraw draws the chosen creature's total toxic value. An
// on-disk key: never renamed, never reused.
var goliathHatcheryDraw = game.RegisterChooseCardsThen("choose-cards/goliath-hatchery-draw",
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		c, ok := g.LookupCardForEffect(r.Picked[0])
		if !ok {
			return nil
		}
		n := game.ToxicTotal(&c)
		if n <= 0 {
			return nil
		}
		return g.DrawNForEffect(r.Chooser, n)
	})
