package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enter the God-Eternals — Sorcery {2}{U}{U}{B} (Starter Commander Decks):
//
//	"Enter the God-Eternals deals 4 damage to target creature and you gain
//	 life equal to the damage dealt this way. Target player mills four
//	 cards. Amass Zombies 4. (...)"
//
// Two target clauses, each re-checked on its own (CR 608.2b), run in printed
// order. The life is the damage actually dealt, through Corrupt's keyed tail
// (a prevented point gains nothing, and a creature that is no longer a legal
// target takes nothing and gains nothing). The mill is the targeted player's,
// who may be anyone, the caster included. Amass is the shared keyword action
// (CR 701.47a) and happens whatever became of the targets, so the Army still
// grows when both are gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67ee7c8b-afcc-4f09-a184-3db01f2ab175",
		Name:         "Enter the God-Eternals",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 4, Targets: game.ForTargets(effects.DamageToTarget(0, 4))},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPlayer("target player"),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return ctx.Game.DealDamageEachKeyedForEffect(ctx.Source(), []uuid.UUID{t.ID}, 4,
					ctx.Tail(effects.GainLifeEqualToDamageDealt))
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetPlayer {
					return nil
				}
				return effects.MillCards{Player: t.ID, N: 4}.Apply(ctx)
			}),
			effects.Amass{Subtype: "Zombie", N: 4},
		),
	})
}
