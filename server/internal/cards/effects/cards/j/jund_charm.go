package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jund Charm — Instant {B}{R}{G} (Shards of Alara, #5588):
//
//	"Choose one —
//	 • Exile target player's graveyard.
//	 • Jund Charm deals 2 damage to each creature.
//	 • Put two +1/+1 counters on target creature."
//
// A modal spell (CR 700.2). The second bullet is Pyroclasm's sweep, one
// damage event to every creature, and declares the sweep. A targeted
// bullet whose target has gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7bd2ac13-d598-4cf1-a483-2a6878aed93b",
		Name:         "Jund Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Exile target player's graveyard.", effects.TargetPlayer("target player"), effects.ExileTheModesTargetPlayersGraveyard),
			effects.ModeWithPurpose(effects.ModeDoing("Jund Charm deals 2 damage to each creature.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DamageEachMatching(ctx, effects.Creature(), 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
			effects.ModeDoing("Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"),
				effects.PlusOneCountersOnTheModesTarget(2)),
		),
	})
}
