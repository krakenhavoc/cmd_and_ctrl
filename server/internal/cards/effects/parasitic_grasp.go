package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parasitic Grasp — Instant {1}{B}:
//
//	"Cleave {1}{B}{B} (You may cast this spell for its cleave cost. If
//	 you do, remove the words in square brackets.)
//	 Parasitic Grasp deals 3 damage to target [Human] creature. You gain
//	 3 life."
//
// Hard-cast it needs a Human; cleaved any creature (CR 702.148a). The
// damage and the life are one spell, so a target that is gone as it
// resolves makes the whole spell do nothing (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "d9763c75-75f7-4766-8cca-84c899796e3c",
		Name:         "Parasitic Grasp",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			LifeGain: 3,
			Targets:  game.ForTargets(DamageToTarget(0, 3)),
		},
		Targets: TargetCreature("target Human creature", Subtype("Human")),
		AlternativeCosts: []game.AlternativeCost{
			Cleave("{1}{B}{B}", TargetCreature("target creature")),
		},
		OnResolve: ResolveSteps(
			// The life waits for a prompt the damage asks, CR 608.2c.
			Instruction(func(ctx *Context) error {
				t, ok := firstTargetCard(ctx.Item)
				if !ok {
					return nil
				}
				return DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}.Apply(ctx)
			}),
			GainLife{Amount: 3},
		),
	})
}
