package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Twisted Fates — Sorcery {2}{W}{W}{B} (Reality Fracture, tracker #2795):
//
//	"Destroy target nonland permanent. Put a +1/+1 counter on each
//	 creature target player controls."
//
// Two target clauses, each judged on its own at resolution (CR 608.2b):
// losing the permanent does not cancel the counters, and losing the
// player does not cancel the destruction. The creatures are read after
// the destruction, so the destroyed permanent gets no counter. The
// counters go on as a batch of ordinary placements.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ce30a19-e716-478f-be89-e2d95753cfed",
		Name:         "Twisted Fates",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetPermanent("target nonland permanent", Nonland()),
			TargetPlayer("target player"),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			p, havePlayer := ctx.ClauseTarget(1)
			havePlayer = havePlayer && p.Kind == game.TargetPlayer
			counters := func(g *game.Game, _ []uuid.UUID) error {
				if !havePlayer {
					return nil
				}
				c := NewContext(g, item)
				var ids []uuid.UUID
				for _, card := range g.BattlefieldCardsForEffect() {
					if card.IsCreature() && card.Controller == p.ID {
						ids = append(ids, card.InstanceID)
					}
				}
				for _, id := range ids {
					if err := (AddCounter{Target: id, Kind: "+1/+1", N: 1}).Apply(c.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			}
			// The counters follow the destruction (printed order), from its
			// continuation, so a replacement that pauses it cannot reorder them.
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				return ctx.Game.DestroyPermanentsThenForEffect([]uuid.UUID{t.ID}, counters)
			}
			return counters(ctx.Game, nil)
		},
	})
}
