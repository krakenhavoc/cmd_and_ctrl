package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7bd2ac13-d598-4cf1-a483-2a6878aed93b",
		Name:         "Jund Charm",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Exile target player's graveyard.", TargetPlayer("target player"), exileTheModesTargetPlayersGraveyard),
			ModeWithPurpose(ModeDoing("Jund Charm deals 2 damage to each creature.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return damageEachMatching(ctx, Creature(), 2)
				}),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}}),
			ModeDoing("Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"),
				PlusOneCountersOnTheModesTarget(2)),
		),
	})
}
