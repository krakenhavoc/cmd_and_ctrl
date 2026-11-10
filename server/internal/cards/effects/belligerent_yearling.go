package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Belligerent Yearling — Creature — Dinosaur {1}{R}, 1/2 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Trample
//	 Whenever another Dinosaur you control enters, you may have this
//	 creature's base power become equal to that creature's power until end
//	 of turn."
//
// An optional trigger on another Dinosaur entering under your control
// (read through the layered types, so a changeling counts). The power is
// the entering creature's as the trigger resolves, from its last known
// information if it has since left (CR 608.2h), and sets the Yearling's
// BASE power at layer 7b until end of turn (CR 613.4b), under counters and
// pumps. A Yearling that has left does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d785e3e-8e19-4eea-b09f-9a30c792dba1",
		Name:            "Belligerent Yearling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.HasSubtypeWord("Dinosaur")
			}, "Belligerent Yearling — its base power becomes that creature's power until end of turn", belligerentYearlingSetPower),
				"Belligerent Yearling — have its base power become that creature's power until end of turn?"),
		},
	})
}

func belligerentYearlingSetPower(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if info, ok := ctx.SourcePermanent(); !ok || info.Left {
		return nil
	}
	if item.Trigger == nil {
		return nil
	}
	entered, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	if !ok {
		return nil
	}
	n := entered.PowerForComparison()
	return untilEndOfTurn(ctx, ctx.Source(), nil,
		fmt.Sprintf("Belligerent Yearling — base power %d until end of turn", n),
		game.SetBasePowerMod(n))
}
