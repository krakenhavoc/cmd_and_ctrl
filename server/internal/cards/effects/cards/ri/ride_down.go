package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ride Down — Instant {R}{W}:
//
//	"Destroy target blocking creature. Creatures that were blocked by
//	 that creature this combat gain trample until end of turn."
//
// The attackers the blocker blocks are read before it is destroyed (the
// destruction removes it from combat), and the set is fixed as the spell
// resolves (CR 611.2c). If the target is illegal on resolution the spell
// does nothing (CR 608.2b), so no creature gains trample. A creature the
// blocker already left (it was destroyed and regenerated, say) still
// counts, because it blocked until then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d136448-7331-4a61-b2ad-cfeb23dfa464",
		Name:         "Ride Down",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target blocking creature", effects.BlockingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			var blocked []uuid.UUID
			if blocker, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
				blocked = blocker.BlockedAttackers()
			}
			// The blocked set is fixed already (CR 611.2c) and does not depend on
			// the destruction, so a blocker that regenerates or is indestructible
			// still gives its attackers trample.
			if len(blocked) > 0 {
				if err := (effects.GrantKeywordUntilEOT{
					Match:    effects.B10OneOf(blocked),
					Keywords: []string{"trample"},
					Label:    "Ride Down — creatures blocked by that creature gain trample",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
