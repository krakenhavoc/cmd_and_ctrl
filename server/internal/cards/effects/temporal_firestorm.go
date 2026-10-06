package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Temporal Firestorm — Sorcery {3}{R}{R}:
//
//	"Kicker {1}{W} and/or {1}{U} (You may pay an additional {1}{W}
//	 and/or {1}{U} as you cast this spell.)
//	 Choose up to X creatures and/or planeswalkers you control, where X
//	 is the number of times this spell was kicked. Those permanents
//	 phase out.
//	 Temporal Firestorm deals 5 damage to each creature and each
//	 planeswalker."
//
// Two kicker costs (CR 702.33b, #2153) that share one clause, so X is
// both counted together (CR 702.33d). The choice is made as the spell
// resolves and is not a target (a hexproof permanent can be chosen),
// so it is ChoosePermanents; the chosen permanents phase out together
// (CR 702.26) and the damage then goes to each creature and planeswalker
// still phased in (phased-out permanents are treated as though they do
// not exist, CR 702.26b), which is the point of the card. Unkicked it
// is a plain five-damage sweep and asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "a0f25ffa-0e53-4e38-8001-d595fd2c2045",
		Name:          "Temporal Firestorm",
		Completeness:  CompletenessFull,
		OptionalCosts: Kickers("{1}{W}", "{1}{U}"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.KickedTimes()
			if x == 0 {
				return temporalFirestormDamage(ctx)
			}
			return ChoosePermanents{
				Question: "Temporal Firestorm — choose up to X creatures and/or planeswalkers you control to phase out",
				Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
					var out []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == of && (c.IsCreature() || c.IsPlaneswalker()) {
							out = append(out, c.InstanceID)
						}
					}
					return out, 0, x
				},
				Then: func(ctx *Context, picked game.PromptedPicks) error {
					if err := (PhaseOut{Targets: picked.Cards()}).Apply(ctx); err != nil {
						return err
					}
					return temporalFirestormDamage(ctx)
				},
			}.Apply(ctx)
		},
	})
}

// temporalFirestormDamage is "5 damage to each creature and each
// planeswalker", as one damage instance.
func temporalFirestormDamage(ctx *Context) error {
	return damageEachMatching(ctx, Or(Creature(), Planeswalker()), 5)
}
