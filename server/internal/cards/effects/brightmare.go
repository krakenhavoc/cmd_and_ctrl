package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "32025108-0dba-40bd-88db-5fcc48e3d054",
		Name:         "Brightmare",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Brightmare — tap up to one target creature; you gain life equal to its power",
					Do(Instruction(brightmareTap), Instruction(brightmareGainLife))),
				TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

func brightmareTap(ctx *Context) error {
	return TapTarget{Target: firstLegalCardTarget(ctx)}.Apply(ctx)
}

func brightmareGainLife(ctx *Context) error {
	id := firstLegalCardTarget(ctx)
	if id == uuid.Nil {
		return nil
	}
	if p := b31CurrentPowerOf(ctx.Game, id); p > 0 {
		return GainLife{Amount: p}.Apply(ctx)
	}
	return nil
}
