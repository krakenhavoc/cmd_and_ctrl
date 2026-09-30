package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// goad.go — #1599's reusable card-facing goad (CR 701.15) shapes,
// built on the marker Alela, Cunning Conqueror already stamps
// (b33Goad, batch33_helpers.go), which the engine ends "until your
// next turn" per goader (game/goad.go, #1598).
// Two shapes cover every printed goad effect that isn't Alela's own
// per-hit-player pick:
//
//	GoadTarget                    a triggered/activated ability's
//	                              Effect for "goad target creature" /
//	                              "you may goad target creature".
//	GoadAllMatching(ctx, match)   "goad each/all creatures <match>" —
//	                              Disrupt Decorum's "you don't
//	                              control".
//
// Both add the effect's controller's goad to every affected card —
// beside any other player's goad already on it (CR 701.15c, #1598) —
// and the engine ends it as that controller's next turn begins
// (CR 701.15a, game/goad.go). Each also schedules ONE "the goad ends"
// delayed trigger listing the cards, exactly as Alela's own goad does;
// since #1598 it ends nothing in this binary and is kept for restore
// points read by an older one (b33ClearListedGoads).

// GoadTarget is the Effect for a targeted "goad target creature"
// ability: read the announced target back (re-checked for CR 608.2b
// legality — a target that left in response does nothing) and goad
// it.
func GoadTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	id, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return goadAndScheduleClear(ctx, []uuid.UUID{id})
}

// GoadAllMatching goads every creature on the battlefield `match`
// accepts, at the moment the effect resolves — Disrupt Decorum's
// "goad all creatures you don't control". One delayed trigger clears
// the whole set at once, same as Alela's per-attack goad.
func GoadAllMatching(ctx *Context, match func(g *game.Game, c game.Card) bool) error {
	g := ctx.Game
	var ids []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := g.Battlefield.Cards[i]
		if c.IsCreature() && match(g, c) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return goadAndScheduleClear(ctx, ids)
}

// goadAndScheduleClear adds the controller's goad to every listed card
// and schedules the legacy "the goad ends" delayed trigger for the
// controller's next upkeep (see the file comment).
func goadAndScheduleClear(ctx *Context, ids []uuid.UUID) error {
	controller := ctx.Controller()
	for _, id := range ids {
		b33Goad(ctx.Game, id, controller)
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "goad — the goad ends",
		Cards:              ids,
		Body:               clearListedGoadsBody,
	}.Apply(ctx)
}
