package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sagittars' Volley — Instant {2}{G} (Ravnica Allegiance):
//
//	"Destroy target creature with flying. Sagittars' Volley deals 1 damage to
//	 each creature with flying your opponents control."
//
// The destroy uses the target; the damage is one instance to each flier an
// opponent controls as the spell resolves (Squall's body, narrowed to your
// opponents), the destroyed creature already gone if it was one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4c4c7d7-744a-4548-adad-a5fc4fb3f586",
		Name:         "Sagittars' Volley",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: only the opponents' creatures with flying.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}},
		Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DamageEachMatching(ctx,
				effects.And(effects.Creature(), effects.HasKeyword("flying"), effects.OpponentControls()), 1)
		},
	})
}
