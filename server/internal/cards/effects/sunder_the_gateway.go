package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sunder the Gateway — Sorcery {1}{W} (March of the Machine):
//
//	"Choose one —
//	 • Destroy target nontoken artifact or enchantment an opponent
//	   controls. Incubate 2.
//	 • Incubate 2, then transform an Incubator token you control."
//
// First bullet: the incubate is its own sentence, so it happens whether
// or not the destroy did anything (an indestructible target, a
// regeneration shield). A target that is gone or illegal at resolution
// fizzles the whole spell (CR 608.2b), so there is no incubate then
// either, exactly as Merciless Repurposing.
//
// Second bullet: "an Incubator token you control" is a CHOICE made as
// the spell resolves, not a target (no hexproof or "can't be the target"
// relevance, and the spell has no target at all in this mode). It is
// made after the new Incubator has entered, so it may be that one or any
// other Incubator the controller has, and it is offered only Incubators
// with their front face up, since a turned-over one is a Phyrexian
// Token and not an Incubator token. A controller with nothing to choose
// is asked nothing and transforms nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "88d49309-e7d7-486d-8696-cd66bec2fe33",
		Name:         "Sunder the Gateway",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target nontoken artifact or enchantment an opponent controls. Incubate 2.",
				TargetPermanent("target nontoken artifact or enchantment an opponent controls",
					Or(Artifact(), Enchantment()), Not(IsTokenPredicate()), OpponentControls())),
			Mode("Incubate 2, then transform an Incubator token you control."),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
					break
				}
				return Incubate{N: 2}.Apply(ctx)
			}
			if ctx.HasMode(1) {
				return Incubate{N: 2, Then: chooseAnIncubatorAndTransformIt}.Apply(ctx)
			}
			return nil
		},
	})
}

// chooseAnIncubatorAndTransformIt is "transform an Incubator token you
// control": the controller picks one of their Incubator tokens as the
// spell resolves and it is turned over (CR 701.27a). A package-level
// func so it captures nothing.
func chooseAnIncubatorAndTransformIt(ctx *Context) error {
	return ChoosePermanents{
		Question: "Transform an Incubator token you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.Battlefield.Cards {
				if c.Controller == of && IsIncubator()(g, of, c) {
					out = append(out, c.InstanceID)
				}
			}
			return out, 1, 1
		},
		Then: TransformPicked,
	}.Apply(ctx)
}
