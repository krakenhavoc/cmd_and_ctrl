package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chef's Kiss — Instant {1}{R}{R}:
//
//	"Gain control of target spell that targets only a single permanent
//	 or player. Copy it, then reselect the targets at random for the
//	 spell and the copy. The new targets can't be you or a permanent you
//	 control."
//
// The target clause is CR 115.9c's "targets only": one different
// object or player, which is still a permanent on the battlefield or a
// player in the game (TargetsOnlyASinglePermanentOrPlayer). A spell
// aimed at the same creature twice qualifies, and each instance of the
// word "target" is then reselected on its own (the 2021-06-18 ruling).
//
// The steal is ADR 0104's (GainControlOfSpellForEffect, a layer-2
// effect pinned to the spell). It does nothing to a spell you already
// control, which is still a legal target ("including a spell you
// control"), and the copy and the reselection happen either way. The
// copy is yours with the spell's targets (CR 707.10), and then both the
// spell and the copy have every target drawn at random from the legal
// targets other than you and your permanents
// (ReselectTargetsAtRandomForEffect, #1815): the draw may land on the
// target already there, and a target with nothing else to be is
// unchanged, even an illegal one (the rulings). Legality is judged for
// the spell's controller, who is you by then, so an opponent's "target
// creature you control" finds nothing and stays put.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d177c04-9aa3-44b8-bbe4-df5de23a4ce6",
		Name:         "Chef's Kiss",
		Completeness: CompletenessFull,
		Targets: TargetSpell("target spell that targets only a single permanent or player",
			TargetsOnlyASinglePermanentOrPlayer()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			you := ctx.Controller()
			ctx.Game.GainControlOfSpellForEffect(ctx.Source(), t.ID, you,
				"Chef's Kiss — gain control of target spell")
			copyID, err := ctx.Game.CopySpellKeepingTargetsForEffect(t.ID, you)
			if err == game.ErrCardNotFound {
				return nil
			}
			if err != nil {
				return err
			}
			for _, id := range []uuid.UUID{t.ID, copyID} {
				err := ctx.Game.ReselectTargetsAtRandomForEffect(game.RandomRetarget{
					ItemID:  id,
					Drawer:  you,
					Source:  ctx.Source(),
					Exclude: you,
				})
				if err != nil && err != game.ErrCardNotFound {
					return err
				}
			}
			return nil
		},
	})
}
