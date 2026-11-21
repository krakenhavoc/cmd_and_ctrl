package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggressive Negotiations — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it and exile that card. Put a +1/+1 counter on up to one target
//	 creature you control."
//
// The revealed-hand pick with an exile (#2115, ADR 0116's 2026-10-05
// amendment): the whole table sees the hand (CR 701.20a), only a
// nonland card may be chosen, and it is exiled, which is not a discard
// (CR 701.9a): no madness, no discard triggers. A hand with no nonland
// card is revealed and nothing is exiled (CR 609.3).
//
// The counter is a second target clause, "up to one", and runs on the
// next line, before the pick is answered. Nothing can observe that:
// state-based actions and triggers wait for the answer (#1289), and
// the counter does not touch the hand. Each clause is re-checked on its
// own (CR 608.2b), so an illegal opponent still leaves the counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c469133e-174d-476b-b135-bbf15e415e72",
		Name:         "Aggressive Negotiations",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetPlayer("target opponent", effects.Opponent()),
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		),
		// The counter waits for the question the choice asks (CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ChooseFromRevealedHand{
					Player: effects.TargetedPlayer(ctx),
					Filter: effects.Nonland(),
					Label:  "nonland card",
					Exile:  true,
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}.Apply(ctx)
			}),
		),
	})
}
