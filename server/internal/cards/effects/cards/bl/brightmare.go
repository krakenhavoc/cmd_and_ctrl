package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brightmare — Creature — Unicorn {2}{W}, 2/3 (Jumpstart Selesnya deck
// request, #4893):
//
//	"When this creature enters, tap up to one target creature. You gain
//	 life equal to that creature's power."
//
// The target is any creature, the Brightmare itself included, or none. The
// tap comes first; then the life is the target's power as the ability
// resolves, after the tap. With no target chosen, or with the target no
// longer legal, nothing happens (CR 608.2b): there is no "that creature"
// to read. A power below zero gains no life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32025108-0dba-40bd-88db-5fcc48e3d054",
		Name:         "Brightmare",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Brightmare — tap up to one target creature; you gain life equal to its power",
					effects.Do(effects.Instruction(brightmareTap), effects.Instruction(brightmareGainLife))),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

func brightmareTap(ctx *effects.Context) error {
	return effects.TapTarget{Target: effects.FirstLegalCardTarget(ctx)}.Apply(ctx)
}

func brightmareGainLife(ctx *effects.Context) error {
	id := effects.FirstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	if p := effects.B31CurrentPowerOf(ctx.Game, id); p > 0 {
		return effects.GainLife{Amount: p}.Apply(ctx)
	}
	return nil
}
