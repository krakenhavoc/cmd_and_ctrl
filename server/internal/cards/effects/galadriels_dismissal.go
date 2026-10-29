package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Galadriel's Dismissal — Instant {W} (The Hobbit Commander, HOC):
//
//	"Kicker {2}{W}
//	 Target creature phases out. If this spell was kicked, each creature
//	 target player controls phases out instead."
//
// The kicked clause is a different clause (#1716, WhenPaid): the kicker is
// announced at CR 601.2b, so a kicked cast names a player and no creature, and
// an unkicked one names a creature and no player. The creatures are read off
// the player as the spell resolves, then all go in ONE PhaseOut call so they
// phase out together with everything attached to them (CR 702.26a, 702.26g);
// a player who is gone by then is a skipped target (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e46d0b57-5eec-4ab6-b9a7-a5c5cc01cc34",
		Name:         "Galadriel's Dismissal",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{2}{W}"),
			TargetPlayer("target player"))},
		Targets: TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if !ctx.WasKicked() {
				return PhaseOut{Targets: legalTargetCards(item, ctx.Game)}.Apply(ctx)
			}
			player, ok := firstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == player && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			return PhaseOut{Targets: ids}.Apply(ctx)
		},
	})
}
