package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Nova — Instant {3}{W}{W} (Mirrodin, S82 OLD, #5606):
//
//	"Exile target attacking creature and all Equipment attached to it."
//
// Kill Shot's attacking-creature target, judged at announce and on
// resolution (CR 608.2b), and Silence the Believers' exile of what is
// attached: the creature and every Equipment attached to it as the spell
// resolves leave together in one exile (CR 400.7), read on the layered
// view, so an Equipment is never left to fall off first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30dc2a1a-c17e-4e95-9dd8-d0e9be1dcd77",
		Name:         "Soul Nova",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			ids := []uuid.UUID{id}
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.IsAttachedTo(id) && c.HasSubtype("Equipment") {
					ids = append(ids, c.InstanceID)
				}
			}
			ctx.Game.ExileCardsForEffect(ids)
			return nil
		},
	})
}
