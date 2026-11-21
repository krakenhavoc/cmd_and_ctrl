package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror of Life Trapping — Artifact {4} (Commander Legends: Battle for
// Baldur's Gate, #5948):
//
//	"Whenever a creature enters, if it was cast, exile it, then return all
//	 other permanent cards exiled with this artifact to the battlefield under
//	 their owners' control."
//
// Any creature entering, whoever controls it. "If it was cast" is an
// intervening "if" (CR 603.4) read off the entering creature's cast
// provenance (CR 400.7d), checked as the trigger fires. The new exile is
// linked to this artifact (Card.ExiledWith); the cards to return are read
// first, so the creature just exiled is not among them, and all of them go
// back together as one entry. A creature that has already left when the trigger
// resolves is not exiled, but the others still return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ec7ae18-b203-49bd-95f9-ad5482459a23",
		Name:         "Mirror of Life Trapping",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, mirrorOfLifeTrappingCastCreature,
				"Mirror of Life Trapping — exile the creature, then return the other cards exiled with this artifact", mirrorOfLifeTrappingTrap),
		},
	})
}

// mirrorOfLifeTrappingCastCreature is the trigger and its intervening "if":
// a creature entered and it was cast.
func mirrorOfLifeTrappingCastCreature(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCreature() && !c.IsToken() && c.Provenance.FromZone != ""
}

func mirrorOfLifeTrappingTrap(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	var others []uuid.UUID
	if ref, ok := ctx.SourceRef(); ok {
		for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
			others = append(others, c.InstanceID)
		}
	}
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return mirrorOfLifeTrappingReturn(ctx, others)
	}
	tail := ctx.Tail(mirrorOfLifeTrappingReturned)
	tail.Carry = others
	return effects.ExileTarget{Target: id, Linked: true, ThenTail: tail}.Apply(ctx)
}

// mirrorOfLifeTrappingReturned is "then return all other permanent cards
// exiled with this artifact", run once the exile has settled (it may have
// stopped for a replacement's prompt). An on-disk key: never renamed, never
// reused.
var mirrorOfLifeTrappingReturned = effects.CardTail("then/mirror-of-life-trapping-return", func(ctx *effects.Context, o game.TailOutcome) error {
	return mirrorOfLifeTrappingReturn(ctx, o.Tail.Carry)
})

func mirrorOfLifeTrappingReturn(ctx *effects.Context, others []uuid.UUID) error {
	if len(others) == 0 {
		return nil
	}
	return effects.ReturnFromExileTogether{Targets: others}.Apply(ctx)
}
