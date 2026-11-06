package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Razor Rings — Instant {1}{W} (Avatar: The Last Airbender Eternal,
// #5075):
//
//	"Razor Rings deals 4 damage to target attacking or blocking creature.
//	 You gain life equal to the excess damage dealt this way."
//
// The excess is CR 120.10's: the damage that landed beyond what was
// lethal, counting damage already marked (the ruling), read after the
// damage is dealt so a prevention shield shrinks it too
// (socB20DealDamageWithExcess, which also honours a deathtouch source).
// A target illegal on resolution is dealt nothing and no life is gained
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb5afbd7-4d21-4449-9204-eeb4ebf6a72a",
		Name:         "Razor Rings",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Targets: game.ForTargets(DamageToTarget(0, 4))},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			excess, err := socB20DealDamageWithExcess(ctx, id, 4)
			if err != nil || excess <= 0 {
				return err
			}
			return GainLife{Amount: excess}.Apply(ctx)
		},
	})
}
