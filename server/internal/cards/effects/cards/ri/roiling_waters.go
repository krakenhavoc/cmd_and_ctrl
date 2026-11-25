package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Roiling Waters — Sorcery {5}{U}{U} (Oath of the Gatewatch, #5545):
//
//	"Return up to two target creatures your opponents control to their
//	 owners' hands. Target player draws two cards."
//
// Two target clauses (Cease's shape): up to two creatures your opponents
// control, then any player, each read from its own slot. The creatures go
// together (a returned commander may stop for its owner's command-zone answer,
// CR 903.9) and the draw follows in the return's keyed tail, so it still
// happens with no creature chosen, or with every chosen creature gone. A
// player target that is no longer legal draws nothing (CR 608.2b); with every
// target illegal the spell does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77fef82d-e8d0-444b-8d76-68e9b533f38b",
		Name:         "Roiling Waters",
		Completeness: effects.CompletenessFull,
		// Slot 0 bounces; slot 1 draws two.
		Purpose: effects.ForTargets(game.TargetPurpose{Slot: 1, Draws: 2}),
		Targets: effects.Clauses(
			effects.TargetCreature("up to two target creatures your opponents control", effects.OpponentControls()).WithCount(0, 2),
			effects.TargetPlayer("target player"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Slot == 0 && t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return ctx.Game.BounceCardsToHandKeyedForEffect(ids, ctx.ItemTail(roilingWatersDraw))
		},
	})
}

// roilingWatersDraw is "Target player draws two cards." after the return,
// keyed (#4253): the key is an on-disk identity, never renamed or reused.
var roilingWatersDraw = effects.CardTail("then/roiling-waters-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	p, ok := ctx.ClauseTarget(1)
	if !ok || p.Kind != game.TargetPlayer {
		return nil
	}
	return effects.DrawCards{Player: p.ID, N: 2}.Apply(ctx)
})
