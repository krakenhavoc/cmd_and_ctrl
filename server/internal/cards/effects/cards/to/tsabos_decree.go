package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tsabo's Decree — Instant {5}{B} (Invasion, #5616):
//
//	"Choose a creature type. Target player reveals their hand and discards all
//	 creature cards of that type. Then destroy all creatures of that type that
//	 player controls. They can't be regenerated."
//
// The type is chosen as the spell resolves (Kindred Dominance's prompt). The
// hand is revealed, every creature card in it of the type is discarded as one
// batch (a changeling is of every type, CR 702.73a), and when the discard has
// finished every creature of the type that player controls is destroyed
// together without regeneration. A target that became illegal leaves the
// spell with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e0de4d9-7a30-4a2b-94ba-8239baad2a0a",
		Name:         "Tsabo's Decree",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Tsabo's Decree — choose a creature type",
				ctx.ItemTail(tsabosDecreeTypeChosen))
			return nil
		},
	})
}

// tsabosDecreeTypeChosen reveals the target's hand and discards its creature
// cards of the chosen type. An on-disk key: never renamed, never reused.
var tsabosDecreeTypeChosen = effects.CreatureTypeChosen("then/tsabos-decree-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	player, ok := tsabosDecreePlayer(ctx)
	if !ok {
		return nil
	}
	ctx.Game.RevealHandForEffect(player)
	p := ctx.Game.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.IsCreature() && c.HasSubtype(t) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ctx.Game.DiscardCardsKeyedForEffect(player, ids, ctx.Source(), effects.WithWord(ctx.ItemTail(tsabosDecreeDestroy), t))
})

// tsabosDecreeDestroy destroys the target's creatures of the chosen type
// once the discard is done. An on-disk key: never renamed, never reused.
var tsabosDecreeDestroy = effects.CardTail("then/tsabos-decree-destroy", func(ctx *effects.Context, o game.TailOutcome) error {
	player, ok := tsabosDecreePlayer(ctx)
	if !ok || o.Tail.Word == "" {
		return nil
	}
	return effects.DestroyAllMatching{
		Match:             effects.And(effects.Creature(), effects.OfCreatureType(o.Tail.Word), effects.ControlledBy(player)),
		CantBeRegenerated: true,
	}.Apply(ctx)
})

func tsabosDecreePlayer(ctx *effects.Context) (uuid.UUID, bool) {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetPlayer {
		return uuid.Nil, false
	}
	return t.ID, true
}
