package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reciprocate — Instant {W}:
//
//	"Exile target creature that dealt damage to you this turn."
//
// The target clause reads the per-turn record of which creature
// objects dealt damage to the caster (#2149): a creature that has
// since been flickered is a new object and not a legal target
// (CR 400.7), and one that dealt damage to somebody else is not one
// either.
func init() {
	Register(Spec{
		OracleID:     "ebdd29c0-2c33-4410-a05c-80ced58c7b81",
		Name:         "Reciprocate",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature that dealt damage to you this turn", DealtDamageToYouThisTurn()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			// Legality is re-checked as the spell resolves (CR 608.2b):
			// a creature flickered in response is gone as a target.
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ExileTarget{Target: id}.Apply(ctx)
		},
	})
}
