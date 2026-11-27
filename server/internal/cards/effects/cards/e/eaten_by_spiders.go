package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eaten by Spiders — Instant {2}{G} (Avacyn Restored, S82 OLD, #5570):
//
//	"Destroy target creature with flying and all Equipment attached to that
//	 creature."
//
// Wing Snare's target, with the Equipment attached to it as the spell
// resolves destroyed in the same simultaneous event (CR 701.8): the pick is
// judged again at resolution (a creature that lost flying is an illegal
// target, CR 608.2b, and the spell does nothing) and the Equipment are those
// attached then, not when it was cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4c04cec-3a0d-4e86-95ed-8d367fe7220d",
		Name:         "Eaten by Spiders",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			ids := []uuid.UUID{target}
			for _, id := range ctx.Game.AttachmentsOf(target) {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.HasSubtype("Equipment") {
					ids = append(ids, id)
				}
			}
			ctx.Game.DestroyPermanentsForEffect(ids)
			return nil
		},
	})
}
