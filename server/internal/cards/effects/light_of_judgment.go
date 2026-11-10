package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Light of Judgment — Instant {4}{R} (Final Fantasy, #4462):
//
//	"Light of Judgment deals 6 damage to target creature. Destroy up to
//	 one Equipment attached to that creature."
//
// The damage is dealt first; the Equipment is chosen as the spell
// resolves (a resolution-time pick of zero or one, not a target), from
// those attached to the creature then. With the creature an illegal
// target the spell does not resolve and destroys nothing (CR 608.2b);
// a creature with no Equipment is a legal target and the pick is empty.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "932177b8-70f8-4908-a48f-fe03204daddc",
		Name:         "Light of Judgment",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(DamageToTarget(0, 6)),
		OnResolve: ResolveSteps(
			// Heated Argument's first sentence: 6 damage to the still-legal target.
			Instruction(socB26HeatedArgumentDamage),
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ChoosePermanents{
					Question: "Light of Judgment — destroy up to one Equipment attached to that creature",
					Keyed:    lightOfJudgmentDestroy,
					Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.HasSubtype("Equipment") && c.IsAttachedTo(id) {
								out = append(out, c.InstanceID)
							}
						}
						return out, 0, 1
					},
				}.Apply(ctx)
			}),
		),
	})
}

// lightOfJudgmentDestroy is the chosen Equipment destroyed, keyed (#4290).
var lightOfJudgmentDestroy = RegisterPermanentsPick("prompt-run/light-of-judgment-destroy", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		if ids := picked.Cards(); len(ids) > 0 {
			ctx.Game.DestroyPermanentsForEffect(ids)
		}
		return nil
	},
})
