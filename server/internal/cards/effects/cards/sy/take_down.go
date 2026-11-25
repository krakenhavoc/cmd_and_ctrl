package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Take Down — Sorcery {G} (Kaladesh, #5539):
//
//	"Choose one —
//	 • Take Down deals 4 damage to target creature with flying.
//	 • Take Down deals 1 damage to each creature with flying."
//
// A two-bullet modal spell (Barbed Lightning's mode shape): the first
// bullet targets a creature with flying, checked again as it resolves
// (CR 608.2b); the second targets nothing and hits every creature with
// flying, the caster's own included (Rolling Temblor's mass damage, over
// "with" instead of "without"). Flying is read through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0925e112-c089-4aa4-9878-6d6fb0a263c8",
		Name:         "Take Down",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Take Down deals 4 damage to target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")), takeDownSingle),
			effects.ModeWithPurpose(effects.ModeDoing("Take Down deals 1 damage to each creature with flying.", nil, takeDownEach),
				// A creature without flying is spared.
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true}}),
		),
	})
}

// takeDownSingle is the first bullet: 4 damage to its own target, if it is
// still legal.
var takeDownSingle = effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
	return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 4}.Apply(ctx)
})

// takeDownEach is the second bullet: 1 damage to each creature with flying.
func takeDownEach(_ *game.StackItem, ctx *effects.Context, _ int) error {
	return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.HasKeyword("flying")), 1)
}
