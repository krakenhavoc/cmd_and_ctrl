package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Integrity // Intervention — split card, oracle
// 12f252b1-d3da-43af-ab31-0c8a6807c3c4 (Ravnica: Clue Edition):
//
//	Integrity {R/W} Instant: "Target creature gets +2/+2 until end of turn."
//	Intervention {2}{R}{W} Instant: "Intervention deals 3 damage to any
//	 target and you gain 3 life."
//
// Two specs, Intervention under "<oracle_id>#1". The life is gained even if
// the damage is prevented; a target that became illegal counters the spell.
//
// No simplification.
func init() {
	const oracle = "12f252b1-d3da-43af-ab31-0c8a6807c3c4"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Integrity",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return pumpAndGrantFirstTarget(2, 2, "Integrity — +2/+2")(ctx.Game, item)
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Intervention",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), LifeGain: 3},
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if legal := ctx.LegalTargets(); len(legal) > 0 {
				if err := (DealDamage{Source: ctx.Source(), Target: legal[0].ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return GainLife{Amount: 3}.Apply(ctx)
		},
	})
}
