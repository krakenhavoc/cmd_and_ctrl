package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rite of Undoing — Instant {4}{U}:
//
//	"Delve. Return target nonland permanent you control and target nonland permanent you don't control to their owners' hands."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// Two clauses, one per printed "target" (#764), each re-checked on its
// own at resolution (CR 608.2b), and the survivors returned as ONE
// simultaneous move, so a trigger watching the bounce sees both. No
// simplification.
func init() {
	Register(Spec{
		OracleID:     "9dbfa026-e364-4111-a03a-e9b1693bc7b7",
		Name:         "Rite of Undoing",
		Completeness: CompletenessFull,
		Delve:        true,
		Targets: Clauses(
			TargetPermanent("target nonland permanent you control", Nonland(), YouControl()),
			TargetPermanent("target nonland permanent you don't control", Nonland(), OpponentControls()),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for slot := 0; slot < 2; slot++ {
				if t, ok := ctx.ClauseTarget(slot); ok && t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) > 0 {
				ctx.Game.BounceCardsToHandForEffect(ids)
			}
			return nil
		},
	})
}
