package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tibalt's Trickery — Instant {1}{R}:
//
//	"Counter target spell. Choose 1, 2, or 3 at random. Its controller
//	 mills that many cards, then exiles cards from the top of their
//	 library until they exile a nonland card with a different name than
//	 that spell. They may cast that card without paying its mana cost.
//	 Then they put the exiled cards on the bottom of their library in a
//	 random order."
//
// The counter is the ordinary CounterTarget; everything after it is
// done TO and BY the countered spell's controller, who may be anyone
// at the table. The number is drawn from the game's keyed random
// stream (an undo or restore redraws the same one), the mill waits for
// its CR 614 window through MillToZoneKeyedForEffect, and the walk is
// Game.ExileUntilCastFreeForEffect, which is cascade's body with the
// hit test passed in: here a nonland card whose name differs from the
// countered spell's, read BEFORE the counter so a spell that is gone
// afterwards still names its exclusion.
//
// The free cast is made during this resolution (CR 608.2g, #4250), as
// cascade's and discover's are: a resolution_cast prompt over the hit
// holds Tibalt's Trickery on the stack while its chooser casts it or
// declines, and a declined hit goes to the bottom with the rest. If the
// target spell is already gone, the spell is countered on resolution
// (CR 608.2b) and none of this happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6623cb85-fb1e-4561-8b21-0977c25096c2",
		Name:         "Tibalt's Trickery",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			target := ctx.Game.StackItemForEffect(stackID)
			if target == nil {
				return nil
			}
			victim := target.Controller
			name := ""
			if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
				name = card.Name
			}
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			n := tibaltsTrickeryCount(ctx)
			// Keyed (#4253 slice 10a): the victim and the countered
			// spell's name ride the mill's tail.
			return ctx.Game.MillToZoneKeyedForEffect(victim, n, game.ZoneGraveyard, game.MillUntil{},
				game.KeyedTail{Ref: tibaltsTrickeryExile, Player: item.Controller, Source: item.SourceCardID,
					Carry: []uuid.UUID{victim}, Word: name})
		},
	})
}

// tibaltsTrickeryExile is "then exiles cards until a nonland card with a
// different name, and may cast it free": Carry[0] is the victim, Word the
// countered spell's name.
var tibaltsTrickeryExile = game.RegisterTail("then/tibalts-trickery-exile-until", func(g *game.Game, o game.TailOutcome) error {
	victim, name := tailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	return g.ExileUntilCastFreeForEffect(victim, o.Tail.Source, func(c game.Card) bool {
		return !c.IsLand() && c.Name != name
	})
})

// tibaltsTrickeryOptions are the three equally likely numbers, as fixed
// IDs because the keyed random stream picks among IDs.
var tibaltsTrickeryOptions = []uuid.UUID{
	uuid.MustParse("00000000-0000-0000-0000-000000000001"),
	uuid.MustParse("00000000-0000-0000-0000-000000000002"),
	uuid.MustParse("00000000-0000-0000-0000-000000000003"),
}

// tibaltsTrickeryCount is "choose 1, 2, or 3 at random".
func tibaltsTrickeryCount(ctx *Context) int {
	picked := ctx.Game.ChooseAtRandomForEffect(randomDraw(ctx), tibaltsTrickeryOptions, 1)
	if len(picked) == 0 {
		return 1
	}
	for i, id := range tibaltsTrickeryOptions {
		if id == picked[0] {
			return i + 1
		}
	}
	return 1
}
