package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Barrage — Sorcery {3}{R} (Dominaria):
//
//	"Kicker—Sacrifice an artifact or Goblin. (You may sacrifice an artifact or
//	 Goblin in addition to any other costs as you cast this spell.)
//	 Goblin Barrage deals 4 damage to target creature. If this spell was
//	 kicked, it also deals 4 damage to target player or planeswalker."
//
// The kicker is a sacrifice paid at cast (CR 702.33a, 601.2h). The second target
// exists only when it was paid: the kicker's own target clause is the printed
// one with the player-or-planeswalker clause appended (CR 702.33d's "choose
// those targets only if kicked", Mind Spiral's shape), chosen as the spell is
// cast. Each clause is judged for itself as the spell resolves (CR 608.2b): a
// creature that has gone does not stop the player taking 4.
//
// No simplification.
func init() {
	kicker := effects.KickerSacrifice("an artifact or Goblin", effects.Or(effects.Artifact(), effects.HasSubtype("Goblin")))
	kicker.Targets = effects.Clauses(
		effects.TargetCreature("target creature"),
		effects.TargetPlayerOrPlaneswalker(),
	)
	effects.Register(effects.Spec{
		OracleID:      "0071723e-388d-4ae7-bc99-c143dd356c63",
		Name:          "Goblin Barrage",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCreature("target creature"),
		Purpose:       effects.ForTargets(effects.DamageToTarget(0, 4)),
		OptionalCosts: []game.AdditionalCost{kicker},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.OldB208DamageClauseTarget(ctx, 0, 4)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.OldB208DamageClauseTarget(ctx, 1, 4)
			}),
		),
	})
}
