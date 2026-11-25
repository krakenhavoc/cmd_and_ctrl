package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blazing Hope — Instant {W} (Rivals of Ixalan, S82 OLD batch 218, #5531):
//
//	"Exile target creature with power greater than or equal to your life total."
//
// The target clause reads the caster's life total and the creature's current
// power through the layers, as the spell is cast and again as it resolves
// (CR 608.2b): if the life total has risen or the power fallen in the meantime
// the target is illegal and the spell is countered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b8d87d15-0ab4-46e3-8258-8ca267f354fd",
		Name:         "Blazing Hope",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetCreature("target creature with power greater than or equal to your life total",
			func(g *game.Game, caster uuid.UUID, c game.Card) bool {
				p := g.PlayerByIDForEffect(caster)
				return p != nil && c.CurrentPower() >= p.Life
			}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileFirstLegalCardTarget(ctx.Game, item)
		},
	})
}
