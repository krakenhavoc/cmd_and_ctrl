package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Corrupt — Sorcery {5}{B} (Urza's Saga):
//
//	"Corrupt deals damage to any target equal to the number of Swamps you
//	 control. You gain life equal to the damage dealt this way."
//
// The Swamps are counted as the spell resolves (CR 608.2h), Cat-Gator's
// count, so the word "Swamps" follows a text change. The life is the
// damage actually dealt, after prevention and doubling, through Rite of
// Consumption's keyed damage-then-total walk: a prevented point gains
// nothing. A target that is no longer legal takes nothing and nobody
// gains life (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a88a1c53-7de3-423a-a95a-24ed5666ab96",
		Name:         "Corrupt",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			n := effects.SwampsControlledBy(ctx.Game, ctx.Controller(), ctx.TextSubtype("Swamp"))
			if len(ts) == 0 || n <= 0 {
				return nil
			}
			return ctx.Game.DealDamageEachKeyedForEffect(ctx.Source(), []uuid.UUID{ts[0].ID}, n,
				ctx.Tail(effects.GainLifeEqualToDamageDealt))
		},
	})
}
