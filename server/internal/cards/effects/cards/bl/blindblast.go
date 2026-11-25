package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blindblast — Instant {2}{R} (Jumpstart, #5510):
//
//	"Blindblast deals 1 damage to target creature. That creature can't block
//	 this turn.
//	 Draw a card."
//
// Rile's damage-and-draw with Blood Aspirant's "can't block this turn" on
// the same target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9abeb38c-2243-424c-bfbc-87aa45b6ea4f",
		Name:         "Blindblast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(blindblastPurpose())},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
				return err
			}
			if err := effects.TargetCreatureCantBlockThisTurn("Blindblast — that creature can't block this turn")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}

// blindblastPurpose is the one entry for the target clause: one damage and
// "can't block this turn".
func blindblastPurpose() game.TargetPurpose {
	e := effects.CantBlockUntilEOT(0, 0, 0)
	e.Damage = 1
	return e
}
