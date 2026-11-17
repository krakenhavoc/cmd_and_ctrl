package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dmuB07HasteSacrificeTail is "It gains haste. Sacrifice it at the beginning
// of the next end step." for the one card a return put onto the battlefield
// (Balduvian Atrocity). The haste has no duration; the sacrifice is the
// delayed trigger of CR 603.7. An on-disk key: never renamed, never reused.
var dmuB07HasteSacrificeTail = TogetherTail("dmu-b07/haste-then-sacrifice",
	func(ctx *Context, entered []uuid.UUID, _ int) error {
		var here []uuid.UUID
		for _, id := range entered {
			if onBattlefield(ctx.Game, id) {
				here = append(here, id)
			}
		}
		if len(here) == 0 {
			return nil
		}
		for _, id := range here {
			if err := (ScopedEffectFor{Target: id, Mods: []game.Mod{game.AddKeywordsMod("haste")},
				Duration: game.IndefiniteDuration(), Label: "Balduvian Atrocity — haste"}).Apply(ctx); err != nil {
				return err
			}
		}
		for _, id := range here {
			c, ok := battlefieldCardForEffect(ctx.Game, id)
			if !ok {
				continue
			}
			if err := (ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Item.Controller,
				Label:      "Balduvian Atrocity — sacrifice it",
				Body:       cmdr2S12SoulstokeSacrifice,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// dmuB07DrawACardBody is "draw a card" as a delayed trigger's body (Warhost's
// Frenzy). An on-disk key: never renamed, never reused.
var dmuB07DrawACardBody = game.SimpleDelayedBody("dmu-b07/draw-a-card", func(g *game.Game, item *game.StackItem) error {
	return DrawCards{N: 1}.Apply(NewContext(g, item))
})
